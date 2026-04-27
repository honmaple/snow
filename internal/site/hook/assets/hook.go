package assets

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"reflect"
	"strings"
	"sync"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/spf13/cast"
)

type (
	AssetsHook struct {
		hook.HookImpl
		ctx *core.Context

		hash        sync.Map
		assets      []*Asset
		assetMap    map[string]bool
		preAssetMap map[string]*Asset
	}
)

func (h *AssetsHook) getHash(file string, w io.Writer) error {
	src, err := h.ctx.Theme.Open(file)
	if err != nil {
		return err
	}
	defer src.Close()

	if _, err := io.Copy(w, src); err != nil {
		return err
	}
	return nil
}

func (h *AssetsHook) getAssetHash(asset *Asset) (string, error) {
	hash := md5.New()
	for _, file := range asset.Files {
		matchedFiles, err := fs.Glob(h.ctx.Theme, file)
		if err != nil {
			return "", err
		}
		for _, matchedFile := range matchedFiles {
			if err := h.getHash(matchedFile, hash); err != nil {
				return "", err
			}
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (h *AssetsHook) collectAsset(asset *Asset) (string, error) {
	if _, ok := h.assetMap[asset.Output]; !ok {
		hash, err := h.getAssetHash(asset)
		if err != nil {
			return "", err
		}

		h.assets = append(h.assets, asset)
		h.assetMap[asset.Output] = true
		h.hash.Store(asset.Output, hash)
		return hash, nil
	}
	hash, _ := h.hash.Load(asset.Output)
	return hash.(string), nil
}

func (h *AssetsHook) BeforeBuild() error {
	return nil
}

// 写入收集的文件
func (h *AssetsHook) AfterBuild() error {
	for _, asset := range h.assets {
		fmt.Println(asset.Output)
	}
	return nil
}

func (h *AssetsHook) HandleInit(set template.TemplateSet) error {
	set.RegisterTag("assets", h.assetsTagParser)
	return nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	preAssetMap := make(map[string]*Asset)

	for name := range ctx.Config.GetStringMap("hooks.assets.option") {
		conf := ctx.Config.Sub("hooks.assets.option." + name)
		if conf == nil {
			continue
		}

		asset := &Asset{
			Files:       conf.GetStringSlice("files"),
			Output:      conf.GetString("output"),
			ShowVersion: conf.GetBool("version"),
		}
		if m := conf.Get("filters"); m != nil {
			switch reflect.TypeOf(m).Kind() {
			case reflect.Slice:
				// - libsass:
				//     path: ""
				// - cssmin:
				asset.Filters = make([]map[string]map[string]any, 0)
				for _, item := range m.([]any) {
					for k, v := range cast.ToStringMap(item) {
						asset.Filters = append(asset.Filters, map[string]map[string]any{
							k: cast.ToStringMap(v),
						})
						break
					}
				}
			case reflect.String:
				// libcass,css
				asset.Filters = make([]map[string]map[string]any, 0)
				for name := range strings.SplitSeq(m.(string), ",") {
					asset.Filters = append(asset.Filters, map[string]map[string]any{
						name: nil,
					})
				}
			}
		}
		preAssetMap[name] = asset
	}

	h := &AssetsHook{
		ctx:         ctx,
		assets:      make([]*Asset, 0),
		assetMap:    make(map[string]bool),
		preAssetMap: preAssetMap,
	}
	return h, nil
}

func init() {
	hook.Register("assets", New)
}
