package database

import (
	"context"
	"errors"
	"testing"

	"kvdb/internal/database/mocks"
	"kvdb/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestDatabase_RunCommand_OK(t *testing.T) {
	tests := []struct {
		name           string
		rawQuery       string
		parseResult    model.Query
		parseError     error
		execResult     string
		execError      error
		expectedOutput string
	}{
		{
			name:     "valid GET command",
			rawQuery: "get key",
			parseResult: model.Query{
				Command: model.CommandGET,
				Args:    []string{"key"},
			},
			parseError:     nil,
			execResult:     "value",
			execError:      nil,
			expectedOutput: "value",
		},
		{
			name:     "valid SET command",
			rawQuery: "set key value",
			parseResult: model.Query{
				Command: model.CommandSET,
				Args:    []string{"key", "value"},
			},
			parseError:     nil,
			execResult:     messageOK,
			execError:      nil,
			expectedOutput: messageOK,
		},
		{
			name:     "valid DEL command",
			rawQuery: "del key",
			parseResult: model.Query{
				Command: model.CommandDEL,
				Args:    []string{"key"},
			},
			parseError:     nil,
			execResult:     messageOK,
			execError:      nil,
			expectedOutput: messageOK,
		},
		{
			name:           "parse error",
			rawQuery:       "invalid query",
			parseResult:    model.Query{},
			parseError:     errors.New("parse error"),
			execResult:     "",
			execError:      nil,
			expectedOutput: "failed parse query: parse error",
		},
		{
			name:     "unknown command",
			rawQuery: "unknown key",
			parseResult: model.Query{
				Command: model.CommandUNK,
				Args:    []string{"key"},
			},
			parseError:     nil,
			execResult:     "",
			execError:      nil,
			expectedOutput: ErrUnknownCommand.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем mock compute
			mockCompute := mocks.NewCompute(t)
			mockCompute.On("Parse", tt.rawQuery).Return(tt.parseResult, tt.parseError)

			// Создаем mock storage
			mockStorage := mocks.NewStorage(t)

			// Создаем логгер
			logger := zap.NewNop()

			// Создаем Database
			db := New(logger, mockCompute, mockStorage)

			// Настраиваем mock storage в зависимости от команды
			switch tt.parseResult.Command {
			case model.CommandGET:
				mockStorage.On("Get", mock.Anything, tt.parseResult.Args[0]).Return(tt.execResult, true)
			case model.CommandSET:
				mockStorage.On("Set", mock.Anything, tt.parseResult.Args[0], tt.parseResult.Args[1]).Return()
			case model.CommandDEL:
				mockStorage.On("Del", mock.Anything, tt.parseResult.Args[0]).Return()
			}

			// Выполняем команду
			output := db.RunCommand(context.Background(), tt.rawQuery)

			// Проверяем результат
			assert.Equal(t, tt.expectedOutput, output, "unexpected output")

			// Проверяем, что моки были вызваны
			mockCompute.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}
