package fileio

import (
	"fmt"
	"os"
	"path"
)

type SegmentsReader struct {
	dir string
}

func NewSegmentsReader(dir string) *SegmentsReader {
	return &SegmentsReader{
		dir: dir,
	}
}

// nextSegment returns next segment name
// or first segment if previousSegment is empty.
func (s *SegmentsReader) nextSegment(previousSegment string) (string, error) {
	files, err := listFiles(s.dir)
	if err != nil {
		return "", fmt.Errorf("failed to list segments: %w", err)
	}

	if len(files) == 0 {
		return "", nil
	}

	if len(previousSegment) == 0 {
		return files[0].Name(), nil
	}

	for fileID, file := range files {
		if file.Name() == previousSegment && fileID+1 < len(files) {
			return files[fileID+1].Name(), nil
		}
	}

	return "", nil
}

func (s *SegmentsReader) ReadNextSegment(previousSegment string) (string, []byte, error) {
	segment, err := s.nextSegment(previousSegment)
	if err != nil {
		return "", []byte{}, fmt.Errorf("failed to find next segment: %w", err)
	}

	if len(segment) == 0 {
		return previousSegment, []byte{}, nil
	}

	segmentData, err := os.ReadFile(path.Join(s.dir, segment))

	if err != nil {
		return segment, segmentData, fmt.Errorf("failed to read segment: %w", err)
	}

	return segment, segmentData, nil
}

func (s *SegmentsReader) LastSegment() (string, error) {
	files, err := listFiles(s.dir)
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		return "", nil
	}

	return files[len(files)-1].Name(), nil
}
