package fileio

import "fmt"

type SegmentsWriter struct {
	dir string
}

func NewSegmentWriter(dir string) *SegmentsWriter {
	return &SegmentsWriter{
		dir: dir,
	}
}

func (s *SegmentsWriter) WriteFullSegment(segment string, data []byte) error {
	file, err := CreateFile(SegmentPath(s.dir, segment))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	if err = file.Sync(); err != nil {
		return err
	}

	return nil
}
