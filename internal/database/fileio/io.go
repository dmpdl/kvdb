package fileio

import (
	"fmt"
	"os"
	"sort"
)

func CreateFile(filename string) (*os.File, error) {
	flags := os.O_CREATE | os.O_WRONLY
	file, err := os.OpenFile(filename, flags, 0644)
	if err != nil {
		return nil, err
	}

	return file, err
}

func WriteFile(file *os.File, data []byte) (int, error) {
	writtenBytes, err := file.Write(data)
	if err != nil {
		return 0, err
	}

	if err = file.Sync(); err != nil {
		return 0, err
	}

	return writtenBytes, nil
}

func listFiles(dir string) ([]os.DirEntry, error) {
	// Create folder if not exists.
	if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("failed to mkdir %s: %w", dir, err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory with segments: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	return files, nil
}
