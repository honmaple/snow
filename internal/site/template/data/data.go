package data

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
	"gopkg.in/yaml.v3"
)

type Data struct {
	ctx    *core.Context
	client *resty.Client
}

func (d *Data) loadFromBytes(data []byte, format string) (any, error) {
	var (
		err    error
		result any
	)
	switch format {
	case "json":
		err = json.Unmarshal(data, &result)
	case "yaml":
		err = yaml.Unmarshal(data, &result)
	default:
		return string(data), nil
	}
	return result, err
}

func (d *Data) loadFromURL(url string, format string) (any, error) {
	req := d.client.R()
	req.SetHeaders(map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5, AppleWebKit/605.1.15 (KHTML, like Gecko,",
	})

	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}
	return d.loadFromBytes(resp.Body(), format)
}

func (d *Data) loadFromFile(path string, format string) (any, error) {
	for _, dir := range []fs.FS{os.DirFS("."), d.ctx.Theme} {
		if _, err := fs.Stat(dir, path); err != nil {
			return nil, err
		}
		data, err := fs.ReadFile(dir, path)
		if err != nil {
			return nil, err
		}
		return d.loadFromBytes(data, format)
	}
	return nil, fmt.Errorf("the %s is not found")
}

func (d *Data) Load(path string, format string) any {
	var (
		err    error
		result any
	)
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		result, err = d.loadFromURL(path, format)
	} else {
		result, err = d.loadFromFile(path, format)
	}
	if err != nil {
		d.ctx.Logger.Warn("load data %s err: %s", path, err.Error())
		return nil
	}
	return result
}

func New(ctx *core.Context) (*Data, error) {
	d := &Data{
		ctx:    ctx,
		client: resty.New(),
	}
	return d, nil
}

func init() {
	template.Register("data", func(ctx *core.Context, set template.TemplateSet) error {
		data, err := New(ctx)
		if err != nil {
			return err
		}
		set.Register("load_data", data.Load)
		return nil
	})
}
