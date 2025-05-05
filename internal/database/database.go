package database

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

const (
	messageOK         = "ok"
	messageEmptyValue = "nil"
)

var (
	ErrUnknownCommand = errors.New("unknown command")
	ErrInvalidArgs    = errors.New("invalid arguments")
)

type Compute interface {
	Parse(query string) (Query, error)
}

type Storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Del(ctx context.Context, key string) error
}

type Database struct {
	logger      *zap.Logger
	compute     Compute
	storage     Storage
	commandsMap map[Command]commandExecFunc
}

type commandExecFunc func(ctx context.Context, query Query) (string, error)

func New(
	logger *zap.Logger,
	compute Compute,
	storage Storage,
) *Database {
	db := &Database{
		logger:  logger,
		compute: compute,
		storage: storage,
	}

	db.commandsMap = map[Command]commandExecFunc{
		CommandGET: db.execGET,
		CommandSET: db.execSET,
		CommandDEL: db.execDEL,
	}

	return db
}

func (db *Database) RunCommand(ctx context.Context, rawQuery string) string {
	zapArgs := []zap.Field{
		zap.String("raw_query", rawQuery),
	}
	db.logger.Debug("run command", zapArgs...)

	query, err := db.compute.Parse(rawQuery)
	if err != nil {
		zapArgs = append(zapArgs, zap.Error(err))
		db.logger.Error("failed parse query", zapArgs...)
		return fmt.Sprintf("failed parse query: %s", err.Error())
	}

	exec, ok := db.commandsMap[query.Command]
	if !ok {
		zapArgs = append(zapArgs, zap.Error(ErrUnknownCommand))
		db.logger.Error("unknown command", zapArgs...)
		return ErrUnknownCommand.Error()
	}

	output, err := exec(ctx, query)
	if err != nil {
		zapArgs = append(zapArgs, zap.Error(err))
		db.logger.Error("failed run query", zapArgs...)
		return fmt.Sprintf("failed run query: %s", err.Error())
	}

	return output
}

func (db *Database) execGET(ctx context.Context, query Query) (string, error) {
	if len(query.Args) != CommandGETArgsLen {
		return "", fmt.Errorf("%w: want %d args", ErrInvalidArgs, CommandGETArgsLen)
	}

	value, err := db.storage.Get(ctx, query.Args[0])
	if err != nil {
		return "", fmt.Errorf("failed exec get: %w", err)
	}

	return value, nil
}

func (db *Database) execSET(ctx context.Context, query Query) (string, error) {
	if len(query.Args) != CommandSETArgsLen {
		return "", fmt.Errorf("%w: want %d args", ErrInvalidArgs, CommandSETArgsLen)
	}

	if err := db.storage.Set(ctx, query.Args[0], query.Args[1]); err != nil {
		return "", fmt.Errorf("failed exec set: %w", err)
	}

	return messageOK, nil
}

func (db *Database) execDEL(ctx context.Context, query Query) (string, error) {
	if len(query.Args) != CommandDELArgsLen {
		return "", fmt.Errorf("%w: want %d args", ErrInvalidArgs, CommandDELArgsLen)
	}

	if err := db.storage.Del(ctx, query.Args[0]); err != nil {
		return "", fmt.Errorf("failed exec set: %w", err)
	}

	return messageOK, nil
}
