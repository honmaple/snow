package assets

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"sync"

	"context"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	AssetsHook struct {
		hook.HookImpl
		mu       sync.Mutex
		ctx      *core.Context
		assetsFS fs.FS

		hash        sync.Map
		assets      []*Asset
		assetMap    map[string]bool
		preAssetMap map[string]*Asset
	}
)

func (h *AssetsHook) getHash(file string, w io.Writer) error {
	src, err := h.assetsFS.Open(file)
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
		matchedFiles, err := fs.Glob(h.assetsFS, file)
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
	h.mu.Lock()
	defer h.mu.Unlock()

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
func (h *AssetsHook) AfterBuild(ctx context.Context, w core.Writer) error {
	for _, asset := range h.assets {
		if err := asset.Execute(ctx, h.assetsFS, w); err != nil {
			return err
		}
	}
	return nil
}

func (h *AssetsHook) HandleTemplateSet(set template.TemplateSet) (template.TemplateSet, error) {
	set.RegisterTag("assets", h.assetsTagParser)
	return set, nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	assetsFS, err := ctx.GetFS("assets", true, false)
	if err != nil {
		return nil, err
	}

	preAssetMap := make(map[string]*Asset)
	for name := range ctx.Config.GetStringMap("hooks.assets.option") {
		conf := ctx.Config.Sub("hooks.assets.option." + name)
		if conf == nil {
			continue
		}

		asset := &Asset{
			Files:       conf.GetStringSlice("files"),
			Output:      conf.GetString("output"),
			ShowVersion: conf.GetBool("show_version"),
		}
		if len(asset.Files) == 0 {
			return nil, fmt.Errorf("hooks.assets.option.%s.files is required", name)
		}
		if asset.Output == "" {
			return nil, fmt.Errorf("hooks.assets.option.%s.output is required", name)
		}
		filters, err := normalizeFilters(withImageFormatOption(conf.Get("filters"), asset.Output))
		if err != nil {
			return nil, fmt.Errorf("hooks.assets.option.%s.filters: %w", name, err)
		}
		compiler, err := normalizeSassCompiler(conf.GetString("sass_compiler"))
		if err != nil {
			return nil, fmt.Errorf("hooks.assets.option.%s.sass_compiler: %w", name, err)
		}
		asset.Filters = filters
		asset.SassCompiler = compiler
		preAssetMap[name] = asset
	}

	h := &AssetsHook{
		ctx:         ctx,
		assetsFS:    assetsFS,
		assets:      make([]*Asset, 0),
		assetMap:    make(map[string]bool),
		preAssetMap: preAssetMap,
	}
	return h, nil
}

func init() {
	hook.Register("assets", New)
}
