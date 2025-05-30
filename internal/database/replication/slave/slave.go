package slave

import (
	"bytes"
	"context"
	"fmt"
	"kvdb/internal/database"
	"kvdb/internal/database/replication"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SegmentsReader defines the interface for reading segments
type SegmentsReader interface {
	LastSegment() (string, error)
}

// SegmentsReader defines the interface for writing segments
type SegmentsWriter interface {
	WriteFullSegment(segment string, data []byte) error
}

// Client defines the interface for communication with master
type Client interface {
	Send(ctx context.Context, data []byte) ([]byte, error)
	Close() error
}

// Slave handles synchronization with master node
type Slave struct {
	prevSegment    string
	segmentsCh     chan database.Segment
	logger         *zap.Logger
	syncInterval   time.Duration
	client         Client
	segmentsReader SegmentsReader
	segmentsWriter SegmentsWriter
	gracefulWg     sync.WaitGroup
}

// New creates a new Slave instance
func New(
	logger *zap.Logger,
	syncInterval time.Duration,
	client Client,
	segmentsReader SegmentsReader,
	segmentsWriter SegmentsWriter,
) *Slave {
	return &Slave{
		logger:         logger,
		segmentsCh:     make(chan database.Segment),
		syncInterval:   syncInterval,
		client:         client,
		segmentsReader: segmentsReader,
		segmentsWriter: segmentsWriter,
	}
}

// Run starts the synchronization process
func (s *Slave) Run(ctx context.Context) {
	s.logger.Info(
		"start replication sync",
		zap.String("sync_interval", s.syncInterval.String()),
	)

	s.gracefulWg.Add(1)
	defer s.gracefulWg.Done()

	ticker := time.NewTicker(s.syncInterval)
	defer func() {
		ticker.Stop()
		if err := s.client.Close(); err != nil {
			s.logger.Error("failed to close client", zap.Error(err))
		}
		close(s.segmentsCh)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.synchronize(ctx); err != nil {
				s.logger.Error("synchronization failed", zap.Error(err))
			}
		}
	}
}

// Wait waits for slave to complete all operations
func (s *Slave) Wait() {
	s.gracefulWg.Wait()
}

// Segments returns the channel for receiving segments
func (s *Slave) Segments() <-chan database.Segment {
	return s.segmentsCh
}

// synchronize performs a single synchronization cycle
func (s *Slave) synchronize(ctx context.Context) error {
	// Get previous segment safely
	prevSegment := s.prevSegment
	if prevSegment == "" {
		currentSegment, err := s.segmentsReader.LastSegment()
		if err != nil {
			return err
		}
		s.prevSegment = currentSegment
		prevSegment = currentSegment
	}

	s.logger.Debug("syncing segment",
		zap.String("previous_segment", prevSegment),
	)

	// Create request
	req := &replication.Request{
		PrevSegment: prevSegment,
	}

	// Send request to master
	response, err := s.client.Send(ctx, replication.Encode(req))
	if err != nil {
		return err
	}

	// Decode response
	var replicationResponse replication.Response
	if err := replication.Decode(&replicationResponse, response); err != nil {
		s.logger.Error("failed to decode response",
			zap.ByteString("raw_response", response),
		)
		return err
	}

	// Check for empty response
	if replicationResponse.Segment == s.prevSegment {
		s.logger.Debug("no new segments available")
		return nil
	}

	// Process logs
	logs, err := s.decodeLogs(replicationResponse.Data)
	if err != nil {
		return err
	}

	// Write new segment to WAL log
	if err := s.segmentsWriter.WriteFullSegment(
		replicationResponse.Segment,
		replicationResponse.Data,
	); err != nil {
		return fmt.Errorf("failed to write segment: %w", err)
	}

	// Send segment to channel
	segment := database.Segment{
		Name: replicationResponse.Segment,
		Logs: logs,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.segmentsCh <- segment:
		s.prevSegment = replicationResponse.Segment
		return nil
	}
}

// decodeLogs decodes WAL records from binary data
func (s *Slave) decodeLogs(data []byte) ([]database.WALRecord, error) {
	buffer := bytes.NewBuffer(data)
	var logs []database.WALRecord

	for buffer.Len() > 0 {
		var log database.WALRecord
		if err := log.Decode(buffer); err != nil {
			s.logger.Error("failed to decode log record",
				zap.Error(err),
				zap.Int("remaining_bytes", buffer.Len()),
			)
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}
