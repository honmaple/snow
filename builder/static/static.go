package static

import (
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

type (
	Static struct {
		URL     string
		File    string
		Root    fs.FS
		IsTheme bool
	}
	Statics []*Static
)

func (s *Static) Name() string {
	if s.IsTheme {
		return filepath.Join("@theme", s.File)
	}
	return s.File
}

func (s *Static) Bytes() ([]byte, error) {
	f, err := s.Root.Open(s.File)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (s *Static) CopyTo(dst string) error {
	buf, err := s.Bytes()
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
