package fileio

import (
	"fmt"
	"os"
)

type SegmentsProcessor struct {
	dir string
}

func NewSegmentProcessor(dir string) *SegmentsProcessor {
	return &SegmentsProcessor{
		dir: dir,
	}
}

func (s *SegmentsProcessor) ForEach(action func([]byte) error) error {
	files, err := listFiles(s.dir)
	if err != nil {
		return fmt.Errorf("failed to list segments: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := fmt.Sprintf("%s/%s", s.dir, file.Name())
		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		if err := action(data); err != nil {
			return err
		}
	}

	return nil
}
