/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: org.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-08-29 14:56:00 (CST)
 Last Update: Friday 2019-09-06 02:47:33 (CST)
		  By:
 Description:
 *********************************************************************************/
package reader

import (
	"bytes"
	"io/ioutil"
	org "org-golang"
	"snow/core"
	"strings"
)

// OrgReader ..
type OrgReader struct {
	Reader
}

func NewOrgReader() *OrgReader {
	reader := new(OrgReader)
	reader.ReaderType = reader
	reader.Extensions = []string{".org"}
	return reader
}

// read ..
func (s *OrgReader) read(file string) (*Metadata, error) {
	var (
		buffer bytes.Buffer
		err    error
	)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	m := &org.Org{
		Block: org.Block{
			Name:      "org",
			NeedParse: true,
		},
	}
	for _, str := range strings.Split(string(data), "\n") {
		m.Append(str)
	}
	for _, str := range m.Children {
		buffer.WriteString(str.HTML())
	}
	metadata, err := s.metadata(org.Meta())
	if err != nil {
		return nil, err
	}
	content := buffer.String()
	metadata.Content = content
	metadata.Summary = s.summary(content)
	return metadata, nil
}

func (s *OrgReader) Read(file string) (*core.Article, error) {
	return s.ReadArticle(file)
}

func (s *OrgReader) ReadArticle(file string) (*core.Article, error) {
	metadata, err := s.read(file)
	if err != nil {
		return nil, err
	}
	article := &core.Article{
		Title:    metadata.Title,
		Summary:  metadata.Summary,
		Content:  metadata.Content,
		Date:     metadata.Date,
		Tags:     metadata.Tags,
		Category: metadata.Category,
		Author:   metadata.Author,
		Authors:  metadata.Authors,
	}
	return article, nil
}
