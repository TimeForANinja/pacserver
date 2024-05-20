package utils

import (
	"os"
	"path/filepath"
)

func ListFiles(root string) ([]string, error) {
	return _listFiles(root, "./")
}

func _listFiles(root, stack string) ([]string, error) {
	full_dir := filepath.Join(root, stack)
	entry, err := os.ReadDir(full_dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)

	for _, e := range entry {
		fullItemPath := filepath.Join(stack, e.Name())
		if e.IsDir() {
			nested, err := _listFiles(root, fullItemPath)
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
