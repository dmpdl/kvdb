package slave

import (
	"context"
	"fmt"
	"kvdb/internal/database/replication"
	"time"

	"go.uber.org/zap"
)

type SegmentsReader interface {
	LastSegment() (string, error)
}

type Client interface {
	Send(context.Context, []byte) ([]byte, error)
	Close() error
}

type Slave struct {
	currentSegment string
	logger         *zap.Logger
	syncInterval   time.Duration
	client         Client
	segmentsReader SegmentsReader
}

func New(
	logger *zap.Logger,
	syncInterval time.Duration,
	client Client,
	segmentsReader SegmentsReader,
) *Slave {
	return &Slave{
		logger:         logger,
		syncInterval:   syncInterval,
		client:         client,
		segmentsReader: segmentsReader,
	}
}

func (s *Slave) Run(ctx context.Context) {
	ticker := time.NewTicker(s.syncInterval)
	defer func() {
		ticker.Stop()
		s.client.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.logger.Debug("sync", zap.String("current_segment", s.currentSegment))
			s.synchronize(ctx)
		}
	}
}

func (s *Slave) synchronize(ctx context.Context) {
	if len(s.currentSegment) == 0 {
		currentSegment, err := s.segmentsReader.LastSegment()
		if err != nil {
			s.logger.Error("failed to get last segment", zap.Error(err))
			return
		}

		s.currentSegment = currentSegment
	}

	response, err := s.client.Send(ctx, replication.Encode(
		&replication.Request{
			PreviousSegment: s.currentSegment,
		},
	))
	if err != nil {
		s.logger.Error("failed to send request to master", zap.Error(err))
		return
	}

	var replicationResponse replication.Response
	if err := replication.Decode(&replicationResponse, response); err != nil {
		s.logger.Error(
			"failed to decode master response",
			zap.Error(err), zap.String("response", string(response)),
		)
		return
	}

	fmt.Println(s.currentSegment)
	s.currentSegment = replicationResponse.Segment
}
