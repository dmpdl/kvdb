package writer

import (
	"bytes"
	"fmt"
	"kvdb/internal/database"
)

type Segment interface {
	Write([]byte) error
}

type LogWriter struct {
	segment Segment
}

func New(segment Segment) *LogWriter {
	return &LogWriter{
		segment: segment,
	}
}

func (lw *LogWriter) Write(requests []database.WriteRequest) error {
	var buffer bytes.Buffer

	for _, request := range requests {
		if err := request.Encode(&buffer); err != nil {
			lw.acknowledgeWrite(
				requests,
				fmt.Errorf("failed encode wal log: %w", err),
			)
			return fmt.Errorf("failed encode wal log: %w", err)
		}
	}

	err := lw.segment.Write(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write wal log: %w", err)
	}

	lw.acknowledgeWrite(requests, nil)

	return nil
}

func (lw *LogWriter) acknowledgeWrite(requests []database.WriteRequest, err error) {
	for _, request := range requests {
		request.SetResponse(err)
	}
}
