package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func ListFiles(path string) ([]string, error) {
	files := make([]string, 0)
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range rd {
		filename := filepath.Join(path, f.Name())
		if !f.IsDir() {
			files = append(files, filename)
			continue
		}
		fs, err := ListFiles(filename)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}
	return files, nil
}

func FileBaseName(file string) string {
	return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
}
