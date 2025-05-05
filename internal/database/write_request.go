package database

import (
	"kvdb/internal/conc"
)

type WriteRequest struct {
	WALRecord
	promise conc.PromiseError
}

func NewWriteRequest(command Command, args []string) WriteRequest {
	return WriteRequest{
		WALRecord: WALRecord{
			Command:   command,
			Arguments: args,
		},
		promise: conc.NewPromise[error](),
	}
}

func (l *WriteRequest) FutureResponse() conc.FutureError {
	return l.promise.GetFuture()
}

func (l *WriteRequest) SetResponse(err error) {
	l.promise.Set(err)
}
