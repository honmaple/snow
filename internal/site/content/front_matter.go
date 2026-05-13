package content

import (
	"github.com/spf13/viper"
)

type FrontMatter struct {
	*viper.Viper
}

func (fm *FrontMatter) Get(key string, defaults ...any) any {
	if fm.IsSet(key) {
		return fm.Viper.Get(key)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

func (fm *FrontMatter) GetBool(key string, defaults ...bool) bool {
	if fm.IsSet(key) {
		return fm.Viper.GetBool(key)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return false
}

func (fm *FrontMatter) GetInt(key string, defaults ...int) int {
	if fm.IsSet(key) {
		return fm.Viper.GetInt(key)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return 0
}

func (fm *FrontMatter) GetInt32(key string, defaults ...int32) int32 {
	if fm.IsSet(key) {
		return fm.Viper.GetInt32(key)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return 0
}

func (fm *FrontMatter) GetInt64(key string, defaults ...int64) int64 {
	if fm.IsSet(key) {
		return fm.Viper.GetInt64(key)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return 0
}

func (fm *FrontMatter) GetString(key string, defaults ...string) string {
	if fm.IsSet(key) {
		return fm.Viper.GetString(key)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}

func (fm *FrontMatter) GetStringSlice(key string, defaults ...[]string) []string {
	if fm.IsSet(key) {
		return fm.Viper.GetStringSlice(key)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

func (fm *FrontMatter) GetStringMap(key string, defaults ...map[string]any) map[string]any {
	if fm.IsSet(key) {
		return fm.Viper.GetStringMap(key)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
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
