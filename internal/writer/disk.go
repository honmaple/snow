package writer

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/honmaple/snow/internal/utils"
)

type DiskWriter struct {
	outputDir string
}

func (w *DiskWriter) WriteFile(ctx context.Context, file string, r io.Reader) error {
	if file == "" {
		return nil
	}
	output := filepath.Join(w.outputDir, filepath.FromSlash(file))

	if dir := filepath.Dir(output); !utils.FileExists(output) {
		os.MkdirAll(dir, 0755)
	}
	dstFile, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, r)
	return err
}

func NewDiskWriter(outputDir string) *DiskWriter {
	return &DiskWriter{outputDir: outputDir}
}
