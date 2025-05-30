package master

import (
	"context"
	"errors"
	"kvdb/internal/database/replication"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestMaster_handleFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		request              []byte
		prepareSegmentReader func(*MockSegmentsReader)
		// mockReaderSetup      func(*MockSegmentsReader)
		// expectedSegment      string
		// expectedData         []byte
		// expectError          bool
		wantResponse []byte
	}{
		{
			name:    "successful segment read",
			request: replication.Encode(&replication.Request{PrevSegment: "seg1"}),
			prepareSegmentReader: func(msr *MockSegmentsReader) {
				msr.EXPECT().ReadNextSegment("seg1").
					Return("seg2", []byte("some data"), nil)
			},
			wantResponse: replication.Encode(&replication.Response{Segment: "seg2", Data: []byte("some data")}),
		},
		{
			name:    "no next segment",
			request: replication.Encode(&replication.Request{PrevSegment: "seg1"}),
			prepareSegmentReader: func(msr *MockSegmentsReader) {
				msr.EXPECT().ReadNextSegment("seg1").
					Return("", nil, nil)
			},
			wantResponse: replication.Encode(&replication.Response{Segment: "seg1"}),
		},
		{
			name:    "reader error",
			request: replication.Encode(&replication.Request{PrevSegment: "seg1"}),
			prepareSegmentReader: func(msr *MockSegmentsReader) {
				msr.EXPECT().ReadNextSegment("seg1").
					Return("", nil, errors.New("some error"))
			},
			wantResponse: replication.ErrResponse(errors.New("failed to read next segment: some error")),
		},
		{
			name:                 "invalid request",
			request:              []byte("invalid data"),
			prepareSegmentReader: func(msr *MockSegmentsReader) {},
			wantResponse:         replication.ErrResponse(errors.New("failed to decode request: unexpected EOF")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t)
			mockReader := NewMockSegmentsReader(t)
			mockServer := NewMockTCPServer(t)

			tt.prepareSegmentReader(mockReader)

			master := New(logger, mockServer, mockReader)

			response := master.handleFunc(context.Background(), tt.request)
			assert.Equal(t, tt.wantResponse, response)
		})
	}
}
