package fileio

import (
	"bytes"
	"kvdb/internal/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegmentProcessor(t *testing.T) {
	sp := NewSegmentProcessor("./test_data")

	// Read logs from segments.
	var logs []database.WALRecord
	sp.ForEach(func(data []byte) error {
		buffer := bytes.NewBuffer(data)

		for buffer.Len() > 0 {
			var log database.WALRecord
			require.NoError(t, log.Decode(buffer))
			logs = append(logs, log)
		}

		return nil
	})

	assert.Equal(t, []database.WALRecord{
		{
			LSN:       1,
			Command:   database.CommandSET,
			Arguments: []string{"key1", "value1"},
		},
		{
			LSN:       3,
			Command:   database.CommandSET,
			Arguments: []string{"key2", "value2"},
		},
		{
			LSN:       4,
			Command:   database.CommandSET,
			Arguments: []string{"key3", "value3"},
		},
		{
			LSN:       5,
			Command:   database.CommandSET,
			Arguments: []string{"key4", "value4"},
		},
		{
			LSN:       6,
			Command:   database.CommandSET,
			Arguments: []string{"key5", "value5"},
		},
		{
			LSN:       7,
			Command:   database.CommandSET,
			Arguments: []string{"key6", "value6"},
		},
	}, logs)
}
