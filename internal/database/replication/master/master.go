package master

import (
	"context"
	"fmt"
	"kvdb/internal/database/replication"

	"go.uber.org/zap"
)

type TCPServer interface {
	ListenFunc(ctx context.Context, f func(ctx context.Context, request []byte) []byte)
}

type SegmentsReader interface {
	ReadNextSegment(previousSegment string) (string, []byte, error)
}

type Master struct {
	logger        *zap.Logger
	tcpServer     TCPServer
	segmentReader SegmentsReader
}

func New(logger *zap.Logger, tcpServer TCPServer, segmentReader SegmentsReader) *Master {
	return &Master{
		logger:        logger,
		tcpServer:     tcpServer,
		segmentReader: segmentReader,
	}
}

func (m *Master) Run(ctx context.Context) {
	m.tcpServer.ListenFunc(ctx, m.handleFunc)
}

func (m *Master) Wait() {}

// handleFunc read next segment and send its data.
// If there is no next segment - send current segment and empty data.
func (m *Master) handleFunc(ctx context.Context, request []byte) []byte {
	var replicationRequest replication.Request
	if err := replication.Decode(&replicationRequest, request); err != nil {
		return replication.ErrResponse(fmt.Errorf("failed to decode request: %w", err))
	}

	segment, segmentData, err := m.segmentReader.ReadNextSegment(replicationRequest.PrevSegment)
	if err != nil {
		m.logger.Error("failed to read next segment", zap.Error(err))
		return replication.ErrResponse(fmt.Errorf("failed to read next segment: %w", err))
	}

	if len(segment) == 0 {
		return replication.Encode(&replication.Response{
			Segment: replicationRequest.PrevSegment,
		})
	}

	return replication.Encode(&replication.Response{
		Segment: segment,
		Data:    segmentData,
	})
}
