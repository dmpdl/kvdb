package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDatabase_RunCommand_OK(t *testing.T) {
	tests := []struct {
		name        string
		rawQuery    string
		parseResult Query
		parseError  error
		execResult  string
		execErr     error
		wantErr     string
		wantResult  string
	}{
		{
			name:     "valid GET command",
			rawQuery: "get key",
			parseResult: Query{
				Command: CommandGET,
				Args:    []string{"key"},
			},
			execResult: "value",
			wantResult: "value",
		},
		{
			name:     "valid SET command",
			rawQuery: "set key value",
			parseResult: Query{
				Command: CommandSET,
				Args:    []string{"key", "value"},
			},
			execResult: messageOK,
			wantResult: messageOK,
		},
		{
			name:     "valid DEL command",
			rawQuery: "del key",
			parseResult: Query{
				Command: CommandDEL,
				Args:    []string{"key"},
			},
			execResult: messageOK,
			wantResult: messageOK,
		},
		{
			name:        "parse error",
			rawQuery:    "invalid query",
			parseResult: Query{},
			parseError:  errors.New("parse error"),
			wantErr:     "parse error",
		},
		{
			name:     "GET storage error",
			rawQuery: "get key",
			parseResult: Query{
				Command: CommandGET,
				Args:    []string{"key"},
			},
			execErr: errors.New("some error"),
			wantErr: "some error",
		},
		{
			name:     "SET storage error",
			rawQuery: "set key value",
			parseResult: Query{
				Command: CommandSET,
				Args:    []string{"key", "value"},
			},
			execErr: errors.New("some error"),
			wantErr: "some error",
		},
		{
			name:     "DEL storage error",
			rawQuery: "del key",
			parseResult: Query{
				Command: CommandDEL,
				Args:    []string{"key"},
			},
			execErr: errors.New("some error"),
			wantErr: "some error",
		},
		{
			name:     "unknown command",
			rawQuery: "unknown key",
			parseResult: Query{
				Command: CommandUNK,
				Args:    []string{"key"},
			},
			wantErr: ErrUnknownCommand.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем mock compute
			mockCompute := NewMockCompute(t)
			mockCompute.On("Parse", tt.rawQuery).Return(tt.parseResult, tt.parseError)

			// Создаем mock storage
			mockStorage := NewMockStorage(t)

			// Создаем логгер
			logger := zap.NewNop()

			// Создаем Database
			db := New(logger, mockCompute, mockStorage)

			// Настраиваем mock storage в зависимости от команды
			switch tt.parseResult.Command {
			case CommandGET:
				mockStorage.EXPECT().Get(mock.Anything, tt.parseResult.Args[0]).Return(tt.execResult, tt.execErr)
			case CommandSET:
				mockStorage.EXPECT().Set(mock.Anything, tt.parseResult.Args[0], tt.parseResult.Args[1]).Return(tt.execErr)
			case CommandDEL:
				mockStorage.EXPECT().Del(mock.Anything, tt.parseResult.Args[0]).Return(tt.execErr)
			}

			// Выполняем команду
			output, err := db.RunCommand(context.Background(), tt.rawQuery)

			// Проверяем результат
			if len(tt.wantErr) != 0 {
				require.Error(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantResult, output, "unexpected output")

			// Проверяем, что моки были вызваны
			mockCompute.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}
