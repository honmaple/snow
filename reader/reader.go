/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: reader.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-08-29 14:52:33 (CST)
 Last Update: Saturday 2019-09-07 17:03:46 (CST)
		  By:
 Description:
 *********************************************************************************/
package reader

import (
	"fmt"
	"path/filepath"
	"snow/core"
	"strings"
	"time"
)

// ReaderType ..
type ReaderType interface {
	extensions() []string
	summary(content string) string
	read(file string) (*Metadata, error)
	metadata(meta map[string]string) (*Metadata, error)

	Read(file string) (*core.Article, error)
	ReadPage(file string) (*core.Page, error)
	ReadArticle(file string) (*core.Article, error)
	ReadStaticFile(file string) error
}

// Metadata ..
type Metadata struct {
	Title    string
	Date     time.Time
	Modified time.Time
	Status   string
	Tags     []string
	Category string
	Author   string
	Authors  []string
	Slug     string
	SaveAs   string
	URL      string
	Content  string
	Summary  string
}

var readers = []ReaderType{
	NewMarkdownReader(),
	NewOrgReader(),
	NewHTMLReader(),
}

type Reader struct {
	ReaderType
	Extensions []string
}

func (s *Reader) extensions() []string {
	return s.Extensions
}

func (s *Reader) metadata(meta map[string]string) (*Metadata, error) {
	metadata := new(Metadata)
	if v, ok := meta["title"]; ok {
		metadata.Title = v
	}
	if v, ok := meta["category"]; ok {
		metadata.Category = v
	}
	if v, ok := meta["author"]; ok {
		metadata.Author = v
	}
	if v, ok := meta["authors"]; ok {
		metadata.Authors = strings.Split(v, ",")
	}
	if v, ok := meta["tags"]; ok {
		metadata.Tags = strings.Split(v, ",")
	}
	if v, ok := meta["date"]; ok {
		date, err := core.ParseTime(v)
		if err != nil {
			return nil, err
		}
		metadata.Date = date
	}
	if v, ok := meta["modify"]; ok {
		date, err := core.ParseTime(v)
		if err != nil {
			return nil, err
		}
		metadata.Modified = date
	}
	if metadata.Title == "" || metadata.Category == "" {
		return nil, fmt.Errorf("title category meta is required")
	}
	return metadata, nil
}

func (s *Reader) summary(content string) string {
	return content
}

func (s *Reader) read(file string) (*Metadata, error) {
	return nil, nil
}

func (s *Reader) Read(file string) (*core.Article, error) {
	if s.ReaderType == nil {
		return s.ReadArticle(file)
	}
	return s.ReaderType.ReadArticle(file)
}

func (s *Reader) ReadPage(file string) (*core.Page, error) {
	metadata, err := s.ReaderType.read(file)
	if err != nil {
		return nil, err
	}
	page := &core.Page{
		Title:   metadata.Title,
		Summary: metadata.Summary,
		Content: metadata.Content,
		Date:    metadata.Date,
	}
	return page, nil
}

func (s *Reader) ReadArticle(file string) (*core.Article, error) {
	metadata, err := s.ReaderType.read(file)
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

func (s *Reader) ReadStaticFile(file string) error {
	return nil
}

func New(extension string) ReaderType {
	if extension == "" {
		return nil
	}
	for _, reader := range readers {
		if core.CheckInList(reader.extensions(), extension) {
			return reader
		}
	}
	return nil
}

func Add(r ReaderType) {
	readers = append(readers, r)
}

func Start() {
	for _, file := range core.G.GetArticles() {
		if reader := New(filepath.Ext(file)); reader != nil {
			if article, err := reader.Read(file); err != nil {
				core.G.Logger.Warn(err.Error())
			} else {
				core.G.AddArticle(article)
			}
		}
	}
	for _, file := range core.G.GetPages() {
		if reader := New(filepath.Ext(file)); reader != nil {
			if page, err := reader.ReadPage(file); err != nil {
				core.G.Logger.Warn(err.Error())
			} else {
				core.G.AddPage(page)
			}
		}
	}
}
