/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: static.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-09-07 00:58:42 (CST)
 Last Update: Saturday 2019-09-07 17:05:25 (CST)
		  By:
 Description:
 *********************************************************************************/
package writer

import (
	"os"
	"path/filepath"
	"snow/core"
)

type StaticWriter struct {
	Writer
}

func NewStaticWriter() *StaticWriter {
	return &StaticWriter{}
}

func (s *StaticWriter) Write() {
	var err error

	for _, static := range core.G.Conf.Statics {
		file := filepath.Join(core.G.Conf.Dir, static.File)
		savePath := static.SavePath
		if savePath == "" {
			savePath = static.File
		}
		savePath = filepath.Join(core.G.Conf.Output, savePath)

		if stat, err := os.Stat(file); err != nil {
			core.G.Logger.Error(err.Error())
			continue
		} else if stat.IsDir() {
			_, err = core.CopyDir(file, savePath)
		} else {
			_, err = core.CopyFile(file, savePath)
		}
		if err != nil {
			core.G.Logger.Error(err.Error())
		}
	}
}
