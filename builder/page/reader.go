package page

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type (
	Builder struct {
		conf    config.Config
		theme   theme.Theme
		hooks   Hooks
		readers map[string]Reader

		pages      Pages
		sections   Sections
		taxonomies Taxonomies
	}
	Reader interface {
		Read(io.Reader) (Meta, error)
	}
)

func (b *Builder) Dirs() []string {
	root := b.conf.GetString("content_dir")
	return []string{root}
}

func (b *Builder) Build(ctx context.Context) error {
	now := time.Now()
	defer func() {
		ls := make([]string, len(b.sections))
		for i, section := range b.sections {
			ls[i] = fmt.Sprintf("%d %s", len(section.Pages), section.Name())
		}
		b.conf.Log.Infoln("Done: Section Processed", strings.Join(ls, ", "), "in", time.Now().Sub(now))

		ts := make([]string, len(b.taxonomies))
		for i, taxonomy := range b.taxonomies {
			ts[i] = fmt.Sprintf("%d %s", len(taxonomy.Terms), taxonomy.Name)
		}
		b.conf.Log.Infoln("Done: Taxonomy Processed", strings.Join(ts, ", "), "in", time.Now().Sub(now))
	}()

	b.loadSections()
	b.loadTaxonomies()
	return b.Write()
}

func NewBuilder(conf config.Config, theme theme.Theme, hooks Hooks) *Builder {
	readers := make(map[string]Reader)
	for ext, c := range _readers {
		readers[ext] = c(conf)
	}
	return &Builder{
		conf:    conf,
		theme:   theme,
		hooks:   hooks,
		readers: readers,
	}
}

type creator func(config.Config) Reader

var _readers = make(map[string]creator)

func Register(ext string, c creator) {
	_readers[ext] = c
}
