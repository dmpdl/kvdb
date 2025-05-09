package wal

import (
	"context"
	"errors"
	"kvdb/internal/conc"
	"kvdb/internal/database"
	"kvdb/internal/database/storage"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockLogsWriter struct {
	t           *testing.T
	mu          sync.Mutex
	recordID    int
	mockRecords [][]database.WALRecord
	mockErr     error
}

func newMockLogsWriter(t *testing.T, batchSize int, records []database.WALRecord, mockErr error) *mockLogsWriter {
	mockRecords := [][]database.WALRecord{}
	batch := []database.WALRecord{}

	for recordID, record := range records {
		batch = append(batch, record)

		if len(batch) == batchSize || recordID == len(records)-1 {
			mockRecords = append(mockRecords, batch)
			batch = []database.WALRecord{}
		}
	}

	mockLogsWriter := &mockLogsWriter{
		t:           t,
		mu:          sync.Mutex{},
		mockRecords: mockRecords,
		mockErr:     mockErr,
	}

	return mockLogsWriter
}

func (m *mockLogsWriter) Write(batch []database.WriteRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.recordID >= len(m.mockRecords) {
		m.t.Error("not enough mocks calls")
	}

	records := make([]database.WALRecord, len(batch))
	for reqID, req := range batch {
		records[reqID] = req.WALRecord
		req.SetResponse(m.mockErr)
	}

	assert.Equal(m.t, m.mockRecords[m.recordID], records)
	m.recordID++

	return m.mockErr
}

func TestNew(t *testing.T) {
	logger := zap.NewNop()
	logsWriter := NewMockLogsWriter(t)
	logsReader := NewMockLogsReader(t)

	w := New(logsWriter, logsReader, logger)

	assert.NotNil(t, w)
	assert.Equal(t, logsWriter, w.logsWriter)
	assert.Equal(t, logsReader, w.logsReader)
}

func TestWithOptions(t *testing.T) {
	w := New(nil, nil, zap.NewNop()).
		WithFlushingBatchLength(10).
		WithFlushingBatchTimeout(100 * time.Millisecond)

	assert.Equal(t, 10, w.opts.flushingBatchLength)
	assert.Equal(t, 100*time.Millisecond, w.opts.flushingBatchTimeout)
}

func TestWAL_SetAndDel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		batchSize   int
		records     []database.WALRecord
		expectError error
	}{
		{
			name:      "ok_full_batch",
			batchSize: 3,
			records: []database.WALRecord{
				{LSN: 1, Command: database.CommandSET, Arguments: []string{"key1", "value1"}},
				{LSN: 2, Command: database.CommandDEL, Arguments: []string{"key2"}},
				{LSN: 3, Command: database.CommandSET, Arguments: []string{"key3", "value3"}},
			},
		},
		{
			name:      "ok_flushing_timeout",
			batchSize: 3,
			records: []database.WALRecord{
				{LSN: 1, Command: database.CommandSET, Arguments: []string{"key1", "value1"}},
				{LSN: 2, Command: database.CommandSET, Arguments: []string{"key2", "value2"}},
			},
		},
		{
			name:      "ok_full_batch_and_flushing_timeout",
			batchSize: 3,
			records: []database.WALRecord{
				{LSN: 1, Command: database.CommandSET, Arguments: []string{"key1", "value1"}},
				{LSN: 2, Command: database.CommandDEL, Arguments: []string{"key2"}},
				{LSN: 3, Command: database.CommandSET, Arguments: []string{"key3", "value3"}},
				{LSN: 4, Command: database.CommandSET, Arguments: []string{"key4", "value4"}},
				{LSN: 5, Command: database.CommandSET, Arguments: []string{"key5", "value5"}},
			},
		},
		{
			name:      "failed_write_batch",
			batchSize: 3,
			records: []database.WALRecord{
				{LSN: 1, Command: database.CommandSET, Arguments: []string{"key1", "value1"}},
				{LSN: 2, Command: database.CommandSET, Arguments: []string{"key2", "value2"}},
				{LSN: 3, Command: database.CommandSET, Arguments: []string{"key3", "value3"}},
			},
			expectError: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockLogsReader := NewMockLogsReader(t)
			mockLogsWriter := newMockLogsWriter(t, tt.batchSize, tt.records, tt.expectError)

			wal := New(mockLogsWriter, mockLogsReader, zap.NewExample()).
				WithFlushingBatchLength(tt.batchSize).
				WithFlushingBatchTimeout(100 * time.Millisecond)

			ctx, cancel := context.WithCancel(context.Background())

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				wal.RunFlushing(ctx)
			}()

			ferrs := []conc.FutureError{}
			for _, r := range tt.records {
				var ferr conc.FutureError

				switch r.Command {
				case database.CommandSET:
					ferr = wal.Set(
						storage.WithTransactionID(context.Background(), r.LSN),
						r.Arguments[0],
						r.Arguments[1],
					)
				case database.CommandDEL:
					ferr = wal.Del(
						storage.WithTransactionID(context.Background(), r.LSN),
						r.Arguments[0],
					)
				}

				ferrs = append(ferrs, ferr)
			}

			for _, ferr := range ferrs {
				require.ErrorIs(t, ferr.Get(), tt.expectError)
			}

			cancel()
			wg.Wait()
		})
	}
}

func TestWAL_Recover(t *testing.T) {
	t.Parallel()

	tErr := errors.New("read failed")

	tests := []struct {
		name          string
		mockRecords   []database.WALRecord
		mockError     error
		expectedError error
	}{
		{
			name: "success",
			mockRecords: []database.WALRecord{
				{LSN: 123, Command: database.CommandSET, Arguments: []string{"key1", "value1"}},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "read error",
			mockRecords:   nil,
			mockError:     tErr,
			expectedError: tErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := new(MockLogsWriter)
			reader := new(MockLogsReader)
			logger := zap.NewNop()

			w := New(writer, reader, logger)

			reader.EXPECT().ReadLogs().Return(tt.mockRecords, tt.mockError)

			records, err := w.Recover()

			require.ErrorIs(t, err, tt.expectedError)
			require.Equal(t, tt.mockRecords, records)
		})
	}
}
