/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: markdown.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-08-29 14:56:42 (CST)
 Last Update: Friday 2019-09-06 03:23:20 (CST)
		  By:
 Description:
 *********************************************************************************/
package reader

import (
	"snow/core"
)

type MarkdownReader struct {
	Reader
}

func NewMarkdownReader() *MarkdownReader {
	reader := new(MarkdownReader)
	reader.ReaderType = reader
	reader.Extensions = []string{".markdown", ".md"}
	return reader
}

func (s *MarkdownReader) ReadArticle(file string) (*core.Article, error) {
	return nil, nil
}
