package query

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		request        []byte
		mockRequest    string
		mockResponse   string
		mockError      error
		expectedOutput []byte
	}{
		{
			name:           "successful command execution",
			request:        []byte("get key"),
			mockRequest:    "get key",
			mockResponse:   "value",
			mockError:      nil,
			expectedOutput: []byte("value"),
		},
		{
			name:           "empty query",
			request:        []byte("   "),
			mockRequest:    "",
			mockResponse:   "",
			mockError:      nil,
			expectedOutput: []byte(""),
		},
		{
			name:           "database error",
			request:        []byte("invalid command"),
			mockRequest:    "invalid command",
			mockResponse:   "",
			mockError:      errors.New("unknown command"),
			expectedOutput: []byte("ERR: unknown command"),
		},
		{
			name:           "query with whitespace",
			request:        []byte("  get key  "),
			mockRequest:    "get key",
			mockResponse:   "value",
			mockError:      nil,
			expectedOutput: []byte("value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockDB := NewMockDatabase(t)
			mockDB.EXPECT().
				RunCommand(mock.Anything, tt.mockRequest).
				Return(tt.mockResponse, tt.mockError)

			// Create handler with mock
			handler := New(mockDB)

			// Execute test
			ctx := context.Background()
			output := handler.Handle(ctx, tt.request)

			// Verify output
			if string(output) != string(tt.expectedOutput) {
				t.Errorf("expected output %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}

func TestNew(t *testing.T) {
	mockDB := NewMockDatabase(t)

	handler := New(mockDB)

	require.NotNil(t, handler)
	require.Equal(t, handler.database, mockDB)
}
