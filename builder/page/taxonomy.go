package page

import (
	"fmt"
	"github.com/honmaple/snow/utils"
	"sort"
	"strings"
)

type (
	TaxonomyConfig struct {
		Path     string `json:"path"`
		Weight   int    `json:"weight"`
		Orderby  string `json:"orderby"`
		Template string `json:"template"`
	}
	Taxonomy struct {
		Path      string
		Permalink string
		Name      string
		Terms     TaxonomyTerms
		Config    TaxonomyConfig
	}
	Taxonomies         []*Taxonomy
	TaxonomyTermConfig struct {
		Path         string      `json:"path"`
		Template     string      `json:"template"`
		Filter       interface{} `json:"filter"`
		Orderby      string      `json:"orderby"`
		Paginate     int         `json:"paginate"`
		FeedPath     string      `json:"feed_path"`
		FeedTemplate string      `json:"feed_template"`
	}
	TaxonomyTerm struct {
		Path      string
		Permalink string
		Name      string
		List      Pages
		Children  TaxonomyTerms
		Config    TaxonomyTermConfig
	}
	TaxonomyTerms []*TaxonomyTerm
)

func (terms TaxonomyTerms) Has(name string) bool {
	for _, term := range terms {
		if term.Name == name {
			return true
		}
	}
	return false
}

func (terms TaxonomyTerms) Orderby(key string) TaxonomyTerms {
	if key == "" {
		return terms
	}
	var (
		reverse = false
		sortf   func(int, int) bool
	)
	if strings.HasSuffix(strings.ToUpper(key), " DESC") {
		key = key[:len(key)-5]
		reverse = true
	}
	switch key {
	case "name":
		sortf = func(i, j int) bool {
			return terms[i].Name < terms[j].Name
		}
	case "count":
		sortf = func(i, j int) bool {
			return len(terms[i].List) < len(terms[j].List)
		}
	}
	if sortf == nil {
		return terms
	}
	if reverse {
		sort.SliceStable(terms, func(i, j int) bool {
			return !sortf(i, j)
		})
	} else {
		sort.SliceStable(terms, sortf)
	}
	return terms
}

func (b *Builder) loadTaxonomyTerms(taxonomy string, terms TaxonomyTerms) {
	for _, term := range terms {
		conf := b.newTaxonomyTermConfig(taxonomy, term.Name)
		term.List = term.List.Filter(conf.Filter).OrderBy(conf.Orderby)
		term.Path = b.conf.GetRelURL(utils.StringReplace(conf.Path, map[string]string{"{number}": "", "{number:one}": "1"}))
		term.Permalink = b.conf.GetURL(term.Path)
		term.Config = conf
		b.loadTaxonomyTerms(taxonomy, term.Children)
	}
}

func (b *Builder) loadTaxonomies() error {
	for name := range b.conf.GetStringMap("taxonomies") {
		if name == "_default" {
			continue
		}
		conf := b.newTaxonomyConfig(name)
		taxonomy := &Taxonomy{
			Name:   name,
			Config: conf,
		}
		taxonomy.Path = b.conf.GetRelURL(utils.StringReplace(conf.Path, map[string]string{"{slug}": b.conf.GetSlug(name)}))
		taxonomy.Permalink = b.conf.GetURL(taxonomy.Path)
		taxonomy.Terms = b.pages.GroupBy(name).Orderby(conf.Orderby)

		b.loadTaxonomyTerms(name, taxonomy.Terms)
		b.taxonomies = append(b.taxonomies, taxonomy)
	}
	sort.SliceStable(b.taxonomies, func(i, j int) bool {
		ti := b.taxonomies[i]
		tj := b.taxonomies[j]
		if ti.Config.Weight == tj.Config.Weight {
			return ti.Name > tj.Name
		}
		return ti.Config.Weight > tj.Config.Weight
	})
	return nil
}

func (b *Builder) newTaxonomyConfig(name string) TaxonomyConfig {
	c := TaxonomyConfig{
		Path:     b.conf.GetString("taxonomies._default.path"),
		Template: b.conf.GetString("taxonomies._default.template"),
	}
	if k := fmt.Sprintf("taxonomies.%s.weight", name); b.conf.IsSet(k) {
		c.Weight = b.conf.GetInt(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.path", name); b.conf.IsSet(k) {
		c.Path = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.orderby", name); b.conf.IsSet(k) {
		c.Orderby = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.template", name); b.conf.IsSet(k) {
		c.Template = b.conf.GetString(k)
	}
	vars := map[string]string{"{taxonomy}": name}
	c.Path = utils.StringReplace(c.Path, vars)
	c.Template = utils.StringReplace(c.Template, vars)
	return c
}

func (b *Builder) newTaxonomyTermConfig(taxonomy, name string) TaxonomyTermConfig {
	c := TaxonomyTermConfig{
		Path:         b.conf.GetString("taxonomies._default.term_path"),
		Template:     b.conf.GetString("taxonomies._default.term_template"),
		FeedPath:     b.conf.GetString("taxonomies._default.feed_path"),
		FeedTemplate: b.conf.GetString("taxonomies._default.feed_template"),
	}
	if k := fmt.Sprintf("taxonomies.%s.term_path", name); b.conf.IsSet(k) {
		c.Path = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.term_template", name); b.conf.IsSet(k) {
		c.Template = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.term_filter", name); b.conf.IsSet(k) {
		c.Filter = b.conf.Get(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.term_orderby", name); b.conf.IsSet(k) {
		c.Orderby = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.term_paginate", name); b.conf.IsSet(k) {
		c.Paginate = b.conf.GetInt(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.feed_path", name); b.conf.IsSet(k) {
		c.FeedPath = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.feed_template", name); b.conf.IsSet(k) {
		c.FeedTemplate = b.conf.GetString(k)
	}
	vars := map[string]string{"{taxonomy}": taxonomy, "{slug}": name}
	c.Path = utils.StringReplace(c.Path, vars)
	c.Template = utils.StringReplace(c.Template, vars)
	c.FeedPath = utils.StringReplace(c.FeedPath, vars)
	c.FeedTemplate = utils.StringReplace(c.FeedTemplate, vars)
	return c
}
