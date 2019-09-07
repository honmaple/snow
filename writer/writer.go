/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: writer.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-04-02 18:15:05 (CST)
 Last Update: Saturday 2019-09-07 16:48:54 (CST)
		  By:
 Description:
 *********************************************************************************/
package writer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"snow/core"
)

// WriterType ..
type WriterType interface {
	Write()
}

// Writer ..
type Writer struct {
}

func (s *Writer) write(file string, content string) {
	if file == "" {
		return
	}
	writefile := filepath.Join(core.G.Conf.Output, file)
	if dir := filepath.Dir(writefile); !core.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}
	core.G.Logger.Debugf("write %s", file)
	if err := ioutil.WriteFile(writefile, []byte(content), 0755); err != nil {
		core.G.Logger.Error(err.Error())
	}
}

var writers = []WriterType{
	NewTemplateWriter(),
	NewFeedWriter(),
	NewStaticWriter(),
}

func Add(w WriterType) {
	writers = append(writers, w)
}

func Start() {
	if len(writers) == 0 {
		core.Exit("no writer", 1)
	}
	for _, writer := range writers {
		writer.Write()
	}
}
