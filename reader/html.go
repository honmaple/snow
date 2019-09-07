/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: html.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-09-04 17:55:57 (CST)
 Last Update: Friday 2019-09-06 02:48:07 (CST)
		  By:
 Description:
 *********************************************************************************/
package reader

import (
	"snow/core"
)

type HTMLReader struct {
	Reader
}

func NewHTMLReader() *HTMLReader {
	reader := new(HTMLReader)
	reader.ReaderType = reader
	reader.Extensions = []string{".html"}
	return reader
}

func (s *HTMLReader) ReadArticle(file string) (*core.Article, error) {
	return nil, nil
}
