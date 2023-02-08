package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

func CopyFile(src, dst string) (written int64, err error) {
	if stat, err := os.Stat(src); err != nil {
		return 0, err
	} else if !stat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return
	}
	defer srcFile.Close()
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dstFile.Close()
	return io.Copy(dstFile, srcFile)
}

func CopyDir(src, dst string) (written int64, err error) {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}
	for _, file := range files {
		srcFile := filepath.Join(src, file.Name())
		dstFile := filepath.Join(dst, file.Name())
		if !file.IsDir() {
			return CopyDir(srcFile, dstFile)
		}
		if written, err = CopyFile(srcFile, dstFile); err != nil {
			return
		}
	}
	return
}

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
