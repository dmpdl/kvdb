package database

import (
	"context"
	"errors"
	"fmt"
	"kvdb/internal/model"

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

//go:generate mockery --name compute --exported --case underscore --with-expecter
type compute interface {
	Parse(query string) (model.Query, error)
}

//go:generate mockery --name storage --exported --case underscore --with-expecter
type storage interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string)
	Del(ctx context.Context, key string)
}

type Database struct {
	logger      *zap.Logger
	compute     compute
	storage     storage
	commandsMap map[model.Command]commandExecFunc
}

type commandExecFunc func(ctx context.Context, query model.Query) (string, error)

func New(
	logger *zap.Logger,
	compute compute,
	storage storage,
) *Database {
	db := &Database{
		logger:  logger,
		compute: compute,
		storage: storage,
	}
	db.commandsMap = map[model.Command]commandExecFunc{
		model.CommandGET: db.execGET,
		model.CommandSET: db.execSET,
		model.CommandDEL: db.execDEL,
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

func (db *Database) execGET(ctx context.Context, query model.Query) (string, error) {
	if len(query.Args) != model.CommandGETArgsLen {
		return "", fmt.Errorf("%w: want %d args", ErrInvalidArgs, model.CommandGETArgsLen)
	}

	value, ok := db.storage.Get(ctx, query.Args[0])
	if !ok {
		return messageEmptyValue, nil
	}

	return value, nil
}

func (db *Database) execSET(ctx context.Context, query model.Query) (string, error) {
	if len(query.Args) != model.CommandSETArgsLen {
		return "", fmt.Errorf("%w: want %d args", ErrInvalidArgs, model.CommandSETArgsLen)
	}

	db.storage.Set(ctx, query.Args[0], query.Args[1])
	return messageOK, nil
}

func (db *Database) execDEL(ctx context.Context, query model.Query) (string, error) {
	if len(query.Args) != model.CommandDELArgsLen {
		return "", fmt.Errorf("%w: want %d args", ErrInvalidArgs, model.CommandDELArgsLen)
	}

	db.storage.Del(ctx, query.Args[0])
	return messageOK, nil
}
