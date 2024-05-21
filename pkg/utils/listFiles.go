package utils

import (
	"io/fs"
	"os"
	"path/filepath"
)

func ListFiles(root string) ([]string, error) {
	return _listFiles(os.ReadDir, root, "./")
}

// some trickery that allows to mock "os" for unit-tests
type readDirFunc func(dirname string) ([]fs.DirEntry, error)

func _listFiles(reader readDirFunc, root, stack string) ([]string, error) {
	full_dir := filepath.Join(root, stack)
	entry, err := reader(full_dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)

	for _, e := range entry {
		fullItemPath := filepath.Join(stack, e.Name())
		if e.IsDir() {
			nested, err := _listFiles(reader, root, fullItemPath)
			if err != nil {
				return nil, err
			}
			files = append(files, nested...)
		} else {
			files = append(files, fullItemPath)
		}
	}

	return files, nil
}
