package utils

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
)

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsExist(err) || err == nil {
		return true
	}
	return false
}

func FileList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, errors.New("is not dir")
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func FileBaseName(file string) string {
	file = filepath.Base(file)
	return file[:len(file)-len(filepath.Ext(file))]
}
