package content

import (
	"github.com/spf13/viper"
)

type FrontMatter struct {
	*viper.Viper
}

func (fm *FrontMatter) MergeFrom(m map[string]any) {
	cf := viper.New()
	for k, v := range m {
		cf.Set(k, v)
	}
	for _, k := range cf.AllKeys() {
		if fm.IsSet(k) {
			continue
		}
		fm.Set(k, cf.Get(k))
	}
}

func NewFrontMatter(m map[string]any) *FrontMatter {
	c := viper.New()
	for k, v := range m {
		c.Set(k, v)
	}
	return &FrontMatter{c}
}
