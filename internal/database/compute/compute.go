package compute

import (
	"errors"
	"fmt"
	"kvdb/internal/database"
	"strings"

	"github.com/google/shlex"
)

var (
	ErrInvalidQuery   = errors.New("invalid query")
	ErrUnknownCommand = errors.New("unknown command")
	ErrInvalidArgs    = errors.New("invalid args")
)

type Compute struct{}

var commandsMap = map[string]database.Command{
	"get": database.CommandGET,
	"set": database.CommandSET,
	"del": database.CommandDEL,
}

const (
	argsLenCommandGet = 1
	argsLenCommandSet = 2
	argsLenCommandDel = 1
)

var argsLenMap = map[database.Command]int{
	database.CommandGET: argsLenCommandGet,
	database.CommandSET: argsLenCommandSet,
	database.CommandDEL: argsLenCommandDel,
}

func New() *Compute {
	return &Compute{}
}

func (c *Compute) Parse(query string) (database.Query, error) {
	queryParts, err := shlex.Split(query)
	if err != nil {
		return database.Query{}, fmt.Errorf("failed to parse query: %w", err)
	}

	if len(queryParts) == 0 {
		return database.Query{}, fmt.Errorf("%w: empty command", ErrInvalidQuery)
	}

	command, ok := mapCommand(queryParts[0])
	if !ok {
		return database.Query{}, fmt.Errorf(
			"%w: unknown command: %s", ErrInvalidQuery, queryParts[0])
	}

	args := queryParts[1:]
	if err := validateArgs(command, args); err != nil {
		return database.Query{}, err
	}

	return database.Query{
		Command: command,
		Args:    args,
	}, nil
}

func mapCommand(commandRaw string) (database.Command, bool) {
	command, ok := commandsMap[strings.ToLower(commandRaw)]
	if !ok {
		return database.CommandUNK, false
	}
	return command, true
}

func validateArgs(command database.Command, args []string) error {
	wantArgsLen, ok := argsLenMap[command]
	if !ok {
		return fmt.Errorf("%w: command %d", ErrUnknownCommand, command)
	}

	if len(args) != wantArgsLen {
		return fmt.Errorf("%w: want %d args %v", ErrInvalidArgs, wantArgsLen, args)
	}

	return nil
}
