package reader

import (
	"bytes"
	"fmt"
	"kvdb/internal/database"
)

type LogsReader struct {
	segmentProcessor SegmentProcessor
}

type SegmentProcessor interface {
	ForEach(func(data []byte) error) error
}

func New(segmentProcessor SegmentProcessor) *LogsReader {
	return &LogsReader{
		segmentProcessor: segmentProcessor,
	}
}

func (lr *LogsReader) ReadLogs() ([]database.WALRecord, error) {
	var logs []database.WALRecord

	err := lr.segmentProcessor.ForEach(func(data []byte) error {
		var err error
		logs, err = lr.readSegment(logs, data)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed read segment: %w", err)
	}

	return logs, err
}

func (lr *LogsReader) readSegment(
	logs []database.WALRecord,
	data []byte,
) ([]database.WALRecord, error) {
	buffer := bytes.NewBuffer(data)

	for buffer.Len() > 0 {
		var log database.WALRecord
		if err := log.Decode(buffer); err != nil {
			return nil, fmt.Errorf("failed to parse logs data: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}
