package compute

import (
	"errors"
	"fmt"
	"kvdb/internal/model"
	"strings"

	"github.com/google/shlex"
)

var (
	ErrInvalidQuery   = errors.New("invalid query")
	ErrUnknownCommand = errors.New("unknown command")
	ErrInvalidArgs    = errors.New("invalid args")
)

type Compute struct{}

var commandsMap = map[string]model.Command{
	"get": model.CommandGET,
	"set": model.CommandSET,
	"del": model.CommandDEL,
}

var argsLenMap = map[model.Command]int{
	model.CommandGET: model.CommandGETArgsLen,
	model.CommandSET: model.CommandSETArgsLen,
	model.CommandDEL: model.CommandDELArgsLen,
}

func New() *Compute {
	return &Compute{}
}

func (c *Compute) Parse(query string) (model.Query, error) {
	queryParts, err := shlex.Split(query)
	if err != nil {
		return model.Query{}, fmt.Errorf("failed to parse query: %w", err)
	}

	if len(queryParts) == 0 {
		return model.Query{}, fmt.Errorf("%w: empty command", ErrInvalidQuery)
	}

	command, ok := mapCommand(queryParts[0])
	if !ok {
		return model.Query{}, fmt.Errorf(
			"%w: unknown command: %s", ErrInvalidQuery, queryParts[0])
	}

	args := queryParts[1:]
	if err := validateArgs(command, args); err != nil {
		return model.Query{}, err
	}

	return model.Query{
		Command: command,
		Args:    args,
	}, nil
}

func mapCommand(commandRaw string) (model.Command, bool) {
	command, ok := commandsMap[strings.ToLower(commandRaw)]
	if !ok {
		return model.CommandUNK, false
	}
	return command, true
}

func validateArgs(command model.Command, args []string) error {
	wantArgsLen, ok := argsLenMap[command]
	if !ok {
		return fmt.Errorf("%w: command %d", ErrUnknownCommand, command)
	}

	if len(args) != wantArgsLen {
		return fmt.Errorf("%w: want %d args %v", ErrInvalidArgs, wantArgsLen, args)
	}

	return nil
}
