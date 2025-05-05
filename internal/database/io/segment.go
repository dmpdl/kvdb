package io

import (
	"fmt"
	"os"
	"time"
)

var now = time.Now

type Segment struct {
	file    *os.File
	size    int
	maxSize int
	dir     string
}

func NewSegment(dir string, maxSize int) *Segment {
	return &Segment{
		dir:     dir,
		maxSize: maxSize,
	}
}

func (s *Segment) Write(data []byte) error {
	if s.file == nil || s.size+len(data) > s.maxSize {
		if err := s.rotateSegment(); err != nil {
			return fmt.Errorf("failed to rotate segment: %w", err)
		}
	}

	writtenBytes, err := WriteFile(s.file, data)
	if err != nil {
		return fmt.Errorf("failed to write data to segment file: %w", err)
	}

	s.size += writtenBytes

	return nil
}

func (s *Segment) rotateSegment() error {
	segmentName := fmt.Sprintf("%s/wal_%d.log", s.dir, now().UnixMilli())
	file, err := CreateFile(segmentName)
	if err != nil {
		return err
	}

	s.file = file
	s.size = 0
	return nil
}
