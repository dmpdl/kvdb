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
	// Create folder if not exists.
	if err := os.Mkdir(s.dir, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed mkdir %s: %w", s.dir, err)
	}

	files, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("failed to scan directory with segments: %w", err)
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
