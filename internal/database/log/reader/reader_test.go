package reader

import (
	"bytes"
	"kvdb/internal/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSegmentProcessor struct {
	logs []database.WALRecord
}

func (sp *mockSegmentProcessor) ForEach(action func(data []byte) error) error {
	var buf bytes.Buffer

	for _, log := range sp.logs {
		log.Encode(&buf)
		if err := action(buf.Bytes()); err != nil {
			return err
		}

		buf.Reset()
	}

	return nil
}

func TestLogsReader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		logs    []database.WALRecord
		wantErr error
	}{
		{
			name: "ok",
			logs: []database.WALRecord{
				{
					LSN:       1,
					Command:   database.CommandSET,
					Arguments: []string{"key1", "value1"},
				},
				{
					LSN:       2,
					Command:   database.CommandSET,
					Arguments: []string{"key2", "value2"},
				},
				{
					LSN:       3,
					Command:   database.CommandDEL,
					Arguments: []string{"key2", "value2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSegmentProcessor := &mockSegmentProcessor{
				logs: tt.logs,
			}

			logsReader := New(mockSegmentProcessor)

			gotLogs, gotErr := logsReader.ReadLogs()
			require.ErrorIs(t, gotErr, tt.wantErr)
			assert.Equal(t, tt.logs, gotLogs)
		})
	}
}
