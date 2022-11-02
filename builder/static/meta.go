package static

import (
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
)

type (
	localFile struct {
		root fs.FS
		file string
	}
	Static struct {
		URL  string
		File interface {
			Name() string
			Bytes() ([]byte, error)
		}
	}
	Statics []*Static
)

func newFile(root fs.FS, file string) *localFile {
	return &localFile{root: root, file: file}
}

func (s *localFile) Name() string {
	return s.file
}

func (s *localFile) Bytes() ([]byte, error) {
	f, err := s.root.Open(s.file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (s *Static) CopyTo(dst string) error {
	buf, err := s.File.Bytes()
	if err != nil {
		return err
	}
	src := bytes.NewBuffer(buf)

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, src)
	return err
}
