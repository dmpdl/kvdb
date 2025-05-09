package storage

import (
	"context"
	"errors"
	"kvdb/internal/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStorage_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		inputKey     string
		outputValue  string
		outputExists bool
		wantErr      error
	}{
		{
			name:         "ok",
			inputKey:     "key1",
			outputValue:  "value1",
			outputExists: true,
		},
		{
			name:         "key not found",
			inputKey:     "key2",
			outputValue:  "",
			outputExists: false,
			wantErr:      ErrKeyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockEngine := NewMockEngine(t)
			mockEngine.EXPECT().
				Get(mock.Anything, tt.inputKey).
				Return(tt.outputValue, tt.outputExists)

			storage, err := New(mockEngine)

			require.NoError(t, err)

			gotValue, gotErr := storage.Get(context.Background(), tt.inputKey)

			require.ErrorIs(t, gotErr, tt.wantErr)
			assert.Equal(t, tt.outputValue, gotValue)
		})
	}
}

func TestStorage_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		inputKey   string
		inputValue string
		wantErr    error
	}{
		{
			name:       "ok",
			inputKey:   "key1",
			inputValue: "value1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockEngine := NewMockEngine(t)
			mockEngine.EXPECT().
				Set(mock.Anything, tt.inputKey, tt.inputValue)

			storage, err := New(mockEngine)

			require.NoError(t, err)

			gotErr := storage.Set(context.Background(), tt.inputKey, tt.inputValue)

			require.ErrorIs(t, gotErr, tt.wantErr)
		})
	}
}

func TestStorage_Del(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		inputKey string
		wantErr  error
	}{
		{
			name:     "ok",
			inputKey: "key1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockEngine := NewMockEngine(t)
			mockEngine.EXPECT().
				Del(mock.Anything, tt.inputKey)

			storage, err := New(mockEngine)

			require.NoError(t, err)

			gotErr := storage.Del(context.Background(), tt.inputKey)

			require.ErrorIs(t, gotErr, tt.wantErr)
		})
	}
}

func TestStorage_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		recoverLogs []database.WALRecord
		recoverErr  error
	}{
		{
			name: "ok",
			recoverLogs: []database.WALRecord{
				{
					LSN:       1,
					Command:   database.CommandSET,
					Arguments: []string{"key1", "value1"},
				},
				{
					LSN:       2,
					Command:   database.CommandDEL,
					Arguments: []string{"key2"},
				},
				{
					LSN:       3,
					Command:   database.CommandDEL,
					Arguments: []string{"key2", "value2"},
				},
			},
		},
		{
			name:        "failed recover",
			recoverLogs: []database.WALRecord{},
			recoverErr:  errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockWAL := NewMockWAL(t)
			mockWAL.EXPECT().Recover().
				Return(tt.recoverLogs, tt.recoverErr)

			mockEngine := NewMockEngine(t)
			for _, log := range tt.recoverLogs {
				switch log.Command {
				case database.CommandSET:
					mockEngine.EXPECT().
						Set(mock.Anything, log.Arguments[0], log.Arguments[1]).
						Return()
				case database.CommandDEL:
					mockEngine.EXPECT().
						Del(mock.Anything, log.Arguments[0]).
						Return()
				}
			}

			_, gotErr := New(mockEngine, WithWAL(mockWAL))

			require.ErrorIs(t, gotErr, tt.recoverErr)
		})
	}
}
