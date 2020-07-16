package extra

import (
	"github.com/honmaple/snow/config"
)

type Extra struct {
	File string
	Path string
}

type Builder struct {
	conf *config.Config
}

func (b *Builder) Read() ([]*Extra, error) {
	extras := make([]*Extra, 0)
	if err := b.conf.UnmarshalKey("extra_dirs", extras); err != nil {
		return nil, err
	}
	return extras, nil
}

func (b *Builder) Build() error {
	extras, err := b.Read()
	if err != nil {
		return err
	}
	return b.Write(extras)
}

func NewBuilder(conf *config.Config) *Builder {
	return &Builder{conf: conf}
}
