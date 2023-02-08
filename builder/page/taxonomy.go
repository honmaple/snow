package page

import (
	"fmt"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
	"sort"
	"strings"
)

type (
	TaxonomyConfig struct {
		Path         string      `json:"path"`
		Weight       int         `json:"weight"`
		Orderby      string      `json:"orderby"`
		Template     string      `json:"template"`
		TermPath     string      `json:"term_path"`
		TermTemplate string      `json:"term_template"`
		TermFilter   interface{} `json:"term_filter"`
		TermOrderby  string      `json:"term_orderby"`
		TermPaginate int         `json:"term_paginate"`
	}
	Taxonomy struct {
		Name   string
		Terms  TaxonomyTerms
		Config TaxonomyConfig
	}
	Taxonomies   []*Taxonomy
	TaxonomyTerm struct {
		Path      string
		Permalink string
		Name      string
		List      Pages
		Children  TaxonomyTerms
	}
	TaxonomyTerms []*TaxonomyTerm
)

func (terms TaxonomyTerms) setURL(conf config.Config, path string) {
	for _, term := range terms {
		term.Path = conf.GetRelURL(utils.StringReplace(path, map[string]string{"{slug}": term.Name}))
		term.Permalink = conf.GetURL(term.Path)
		term.Children.setURL(conf, path)
	}
}

func (terms TaxonomyTerms) Has(name string) bool {
	for _, term := range terms {
		if term.Name == name {
			return true
		}
	}
	return false
}

func (terms TaxonomyTerms) Orderby(key string) TaxonomyTerms {
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

func (b *Builder) loadTaxonomies() error {
	for name := range b.conf.GetStringMap("taxonomies") {
		if name == "_default" {
			continue
		}
		taxonomy := &Taxonomy{
			Name:   name,
			Terms:  b.pages.GroupBy(name),
			Config: b.newTaxonomyConfig(name),
		}
		taxonomy.Terms.setURL(b.conf, taxonomy.Config.TermPath)
		// fmt.Println(name, len(taxonomy.Terms))
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
		Path:         b.conf.GetString("taxonomies._default.path"),
		Template:     b.conf.GetString("taxonomies._default.template"),
		TermPath:     b.conf.GetString("taxonomies._default.term_path"),
		TermTemplate: b.conf.GetString("taxonomies._default.term_template"),
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
	if k := fmt.Sprintf("taxonomies.%s.term_path", name); b.conf.IsSet(k) {
		c.TermPath = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.term_template", name); b.conf.IsSet(k) {
		c.TermTemplate = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.term_filter", name); b.conf.IsSet(k) {
		c.TermFilter = b.conf.Get(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.term_orderby", name); b.conf.IsSet(k) {
		c.TermOrderby = b.conf.GetString(k)
	}
	if k := fmt.Sprintf("taxonomies.%s.term_paginate", name); b.conf.IsSet(k) {
		c.TermPaginate = b.conf.GetInt(k)
	}
	c.Path = utils.StringReplace(c.Path, map[string]string{"{taxonomy}": name})
	c.Template = utils.StringReplace(c.Template, map[string]string{"{taxonomy}": name})
	c.TermPath = utils.StringReplace(c.TermPath, map[string]string{"{taxonomy}": name})
	c.TermTemplate = utils.StringReplace(c.TermTemplate, map[string]string{"{taxonomy}": name})
	return c
}
