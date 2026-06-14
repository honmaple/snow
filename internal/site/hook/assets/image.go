package assets

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	stdpath "path"
	"reflect"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/spf13/cast"
)

const (
	assetFilterImage = "image"

	imageFitInside = "inside"
	imageFitCover  = "cover"
	imageFitFill   = "fill"
)

type ImageFilter struct {
	Width   int
	Height  int
	Fit     string
	Quality int
	Format  string
}

func (f *ImageFilter) Name() string {
	return assetFilterImage
}

func (f *ImageFilter) Execute(buf []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("image filter requires an image input: %w", err)
	}
	img = f.Apply(img)

	var b bytes.Buffer
	if err := f.Encode(&b, img, f.Format); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (f ImageFilter) Apply(img image.Image) image.Image {
	width := f.Width
	height := f.Height
	if width == 0 && height == 0 {
		return img
	}
	if width == 0 || height == 0 {
		bounds := img.Bounds()
		sourceWidth := bounds.Dx()
		sourceHeight := bounds.Dy()
		if width == 0 {
			width = height * sourceWidth / sourceHeight
		}
		if height == 0 {
			height = width * sourceHeight / sourceWidth
		}
	}

	switch f.Fit {
	case imageFitCover:
		return imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)
	case imageFitFill:
		return imaging.Resize(img, width, height, imaging.Lanczos)
	default:
		return imaging.Fit(img, width, height, imaging.Lanczos)
	}
}

func (f ImageFilter) Encode(w io.Writer, img image.Image, format string) error {
	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: f.Quality})
	case "png":
		return png.Encode(w, img)
	case "gif":
		return gif.Encode(w, img, nil)
	default:
		return fmt.Errorf("unsupported image format %q", format)
	}
}

func (n *Asset) hasImageFilter() bool {
	for _, filter := range n.Filters {
		if filter.Name() == assetFilterImage {
			return true
		}
	}
	return false
}

func (n *Asset) executeImage(assetsFS fs.FS) ([]byte, error) {
	matchedFiles, err := n.matchedFiles(assetsFS)
	if err != nil {
		return nil, err
	}
	if len(matchedFiles) != 1 {
		return nil, fmt.Errorf("image filter requires exactly one matched file, got %d", len(matchedFiles))
	}

	buf, err := fs.ReadFile(assetsFS, matchedFiles[0])
	if err != nil {
		return nil, err
	}
	for _, filter := range n.Filters {
		if filter.Name() != assetFilterImage {
			return nil, fmt.Errorf("assets filter %q cannot be combined with image filter", filter.Name())
		}
		buf, err = filter.Execute(buf)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func withImageFormatOption(value any, output string) any {
	format := strings.TrimPrefix(stdpath.Ext(output), ".")
	if format == "" || value == nil {
		return value
	}

	switch filters := value.(type) {
	case []any:
		results := make([]any, 0, len(filters))
		for _, filter := range filters {
			results = append(results, withImageFormatFilter(filter, format))
		}
		return results
	default:
		if reflect.TypeOf(value).Kind() == reflect.Slice {
			values := reflect.ValueOf(value)
			results := make([]any, 0, values.Len())
			for i := 0; i < values.Len(); i++ {
				results = append(results, withImageFormatFilter(values.Index(i).Interface(), format))
			}
			return results
		}
		return value
	}
}

func withImageFormatFilter(value any, format string) any {
	switch filter := value.(type) {
	case map[string]any:
		return withImageFormatMap(filter, format)
	case map[any]any:
		options := make(map[string]any, len(filter))
		for key, value := range filter {
			options[cast.ToString(key)] = value
		}
		return withImageFormatMap(options, format)
	default:
		return value
	}
}

func withImageFormatMap(filter map[string]any, format string) map[string]any {
	name := strings.TrimSpace(cast.ToString(filter["name"]))
	if name != assetFilterImage {
		return filter
	}

	options := make(map[string]any, len(filter)+1)
	for key, value := range filter {
		options[key] = value
	}
	if _, ok := options["format"]; !ok {
		options["format"] = format
	}
	return options
}

func newImageFilter(options map[string]any) (*ImageFilter, error) {
	filter := &ImageFilter{
		Fit:     imageFitInside,
		Quality: 85,
	}
	for key, value := range options {
		switch key {
		case "width", "height":
			size := cast.ToInt(value)
			if size < 0 {
				return nil, fmt.Errorf("image filter %s must be greater than or equal to 0", key)
			}
			if key == "width" {
				filter.Width = size
			} else {
				filter.Height = size
			}
		case "fit":
			fit := strings.TrimSpace(cast.ToString(value))
			switch fit {
			case "", imageFitInside:
				filter.Fit = imageFitInside
			case imageFitCover, imageFitFill:
				filter.Fit = fit
			default:
				return nil, fmt.Errorf("unknown image fit %q", fit)
			}
		case "quality":
			quality := cast.ToInt(value)
			if quality < 1 || quality > 100 {
				return nil, fmt.Errorf("image filter quality must be between 1 and 100")
			}
			filter.Quality = quality
		case "format":
			filter.Format = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(cast.ToString(value))), ".")
		default:
			return nil, fmt.Errorf("unknown image filter option %q", key)
		}
	}
	if (filter.Fit == imageFitCover || filter.Fit == imageFitFill) && (filter.Width == 0 || filter.Height == 0) {
		return nil, fmt.Errorf("image filter fit %q requires width and height", filter.Fit)
	}
	return filter, nil
}
