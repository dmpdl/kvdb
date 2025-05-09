package writer

import (
	"bytes"
	"errors"
	"kvdb/internal/database"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		requests   []database.WriteRequest
		writeError error
	}{
		{
			name: "without error",
			requests: []database.WriteRequest{
				database.NewWriteRequest(12, database.CommandSET, []string{"key", "value"}),
				database.NewWriteRequest(12, database.CommandDEL, []string{"key"}),
			},
		},
		{
			name: "with error",
			requests: []database.WriteRequest{
				database.NewWriteRequest(12, database.CommandSET, []string{"key", "value"}),
				database.NewWriteRequest(12, database.CommandDEL, []string{"key"}),
			},
			writeError: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockSegment := NewMockSegment(t)

			var buffer bytes.Buffer
			for _, request := range tt.requests {
				request.Encode(&buffer)
			}

			mockSegment.EXPECT().Write(buffer.Bytes()).
				Return(tt.writeError)

			lw := New(mockSegment)

			err := lw.Write(tt.requests)

			if tt.writeError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.writeError)
			}
		})
	}
}
