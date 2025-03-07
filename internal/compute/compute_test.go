package compute

import (
	"errors"
	"kvdb/internal/model"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expected    model.Query
		expectedErr error
	}{
		{
			name:  "valid GET command",
			query: `get key`,
			expected: model.Query{
				Command: model.CommandGET,
				Args:    []string{"key"},
			},
			expectedErr: nil,
		},
		{
			name:  "valid SET command",
			query: `set key value`,
			expected: model.Query{
				Command: model.CommandSET,
				Args:    []string{"key", "value"},
			},
			expectedErr: nil,
		},
		{
			name:  "valid DEL command",
			query: `del key`,
			expected: model.Query{
				Command: model.CommandDEL,
				Args:    []string{"key"},
			},
			expectedErr: nil,
		},
		{
			name:        "empty command",
			query:       ``,
			expected:    model.Query{},
			expectedErr: ErrInvalidQuery,
		},
		{
			name:        "unknown command",
			query:       `unknown key`,
			expected:    model.Query{},
			expectedErr: ErrInvalidQuery,
		},
		{
			name:        "invalid GET args",
			query:       `get key extra`,
			expected:    model.Query{},
			expectedErr: ErrInvalidArgs,
		},
		{
			name:        "invalid SET args",
			query:       `set key`,
			expected:    model.Query{},
			expectedErr: ErrInvalidArgs,
		},
		{
			name:        "invalid DEL args",
			query:       `del key extra`,
			expected:    model.Query{},
			expectedErr: ErrInvalidArgs,
		},
	}

	c := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.Parse(tt.query)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if result.Command != tt.expected.Command {
				t.Errorf("expected command: %v, got: %v", tt.expected.Command, result.Command)
			}

			if len(result.Args) != len(tt.expected.Args) {
				t.Errorf("expected args length: %v, got: %v", len(tt.expected.Args), len(result.Args))
			}

			for i, arg := range result.Args {
				if arg != tt.expected.Args[i] {
					t.Errorf("expected arg %d: %v, got: %v", i, tt.expected.Args[i], arg)
				}
			}
		})
	}
}

func TestMapCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.Command
		ok       bool
	}{
		{
			name:     "valid GET command",
			input:    "get",
			expected: model.CommandGET,
			ok:       true,
		},
		{
			name:     "valid SET command",
			input:    "set",
			expected: model.CommandSET,
			ok:       true,
		},
		{
			name:     "valid DEL command",
			input:    "del",
			expected: model.CommandDEL,
			ok:       true,
		},
		{
			name:     "unknown command",
			input:    "unknown",
			expected: model.CommandUNK,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := mapCommand(tt.input)

			if result != tt.expected {
				t.Errorf("expected command: %v, got: %v", tt.expected, result)
			}

			if ok != tt.ok {
				t.Errorf("expected ok: %v, got: %v", tt.ok, ok)
			}
		})
	}
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name        string
		command     model.Command
		args        []string
		expectedErr error
	}{
		{
			name:        "valid GET args",
			command:     model.CommandGET,
			args:        []string{"key"},
			expectedErr: nil,
		},
		{
			name:        "invalid GET args",
			command:     model.CommandGET,
			args:        []string{"key", "extra"},
			expectedErr: ErrInvalidArgs,
		},
		{
			name:        "valid SET args",
			command:     model.CommandSET,
			args:        []string{"key", "value"},
			expectedErr: nil,
		},
		{
			name:        "invalid SET args",
			command:     model.CommandSET,
			args:        []string{"key"},
			expectedErr: ErrInvalidArgs,
		},
		{
			name:        "valid DEL args",
			command:     model.CommandDEL,
			args:        []string{"key"},
			expectedErr: nil,
		},
		{
			name:        "invalid DEL args",
			command:     model.CommandDEL,
			args:        []string{"key", "extra"},
			expectedErr: ErrInvalidArgs,
		},
		{
			name:        "unknown command",
			command:     model.CommandUNK,
			args:        []string{"key"},
			expectedErr: ErrUnknownCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgs(tt.command, tt.args)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}
