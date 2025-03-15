package inmemory

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage_Get(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		data           map[string]string
		expectedValue  string
		expectedExists bool
	}{
		{
			name:           "key exists",
			key:            "key1",
			data:           map[string]string{"key1": "value1"},
			expectedValue:  "value1",
			expectedExists: true,
		},
		{
			name:           "key does not exist",
			key:            "key2",
			data:           map[string]string{"key1": "value1"},
			expectedValue:  "",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем хранилище с тестовыми данными
			storage := &Storage{
				mu:   sync.RWMutex{},
				data: tt.data,
			}

			// Выполняем Get
			value, exists := storage.Get(context.Background(), tt.key)

			// Проверяем результат
			assert.Equal(t, tt.expectedValue, value, "unexpected value")
			assert.Equal(t, tt.expectedExists, exists, "unexpected exists")
		})
	}
}

func TestStorage_Set(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		initialData  map[string]string
		expectedData map[string]string
	}{
		{
			name:         "set new key",
			key:          "key1",
			value:        "value1",
			initialData:  map[string]string{},
			expectedData: map[string]string{"key1": "value1"},
		},
		{
			name:         "update existing key",
			key:          "key1",
			value:        "new_value",
			initialData:  map[string]string{"key1": "value1"},
			expectedData: map[string]string{"key1": "new_value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем хранилище с тестовыми данными
			storage := &Storage{
				mu:   sync.RWMutex{},
				data: tt.initialData,
			}

			// Выполняем Set
			storage.Set(context.Background(), tt.key, tt.value)

			// Проверяем, что данные обновились
			assert.Equal(t, tt.expectedData, storage.data, "unexpected data")
		})
	}
}

func TestStorage_Del(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		initialData  map[string]string
		expectedData map[string]string
	}{
		{
			name:         "delete existing key",
			key:          "key1",
			initialData:  map[string]string{"key1": "value1"},
			expectedData: map[string]string{},
		},
		{
			name:         "delete non-existing key",
			key:          "key2",
			initialData:  map[string]string{"key1": "value1"},
			expectedData: map[string]string{"key1": "value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем хранилище с тестовыми данными
			storage := &Storage{
				mu:   sync.RWMutex{},
				data: tt.initialData,
			}

			// Выполняем Del
			storage.Del(context.Background(), tt.key)

			// Проверяем, что данные обновились
			assert.Equal(t, tt.expectedData, storage.data, "unexpected data")
		})
	}
}
