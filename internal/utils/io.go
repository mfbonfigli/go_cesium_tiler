package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func CreateDirectoryIfDoesNotExist(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err := os.MkdirAll(directory, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func TouchFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}

func FindLasFilesInFolder(directory string) ([]string, error) {
	if _, err := os.Stat(directory); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lastIndex := -1
		name := e.Name()
		if lastIndex = strings.LastIndex(name, "."); lastIndex != -1 {
			ext := e.Name()[lastIndex+1:]
			if strings.ToLower(ext) != "las" {
				continue
			}
		}
		f := filepath.Join(directory, e.Name())
		files = append(files, f)
	}
	return files, nil
}
