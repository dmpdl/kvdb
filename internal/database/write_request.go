package database

import (
	"kvdb/internal/conc"
)

type WriteRequest struct {
	command Command
	args    []string
	promise conc.PromiseError
}

func NewWriteRequest(command Command, args []string) WriteRequest {
	return WriteRequest{
		command: command,
		args:    args,
		promise: conc.NewPromise[error](),
	}
}

func (l *WriteRequest) FutureResponse() conc.FutureError {
	return l.promise.GetFuture()
}
