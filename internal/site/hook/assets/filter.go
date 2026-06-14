package assets

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

func normalizeFilters(value any) ([]Filter, error) {
	if value == nil {
		return nil, nil
	}

	var names []string

	switch filters := value.(type) {
	case string:
		names = strings.Split(filters, ",")
	case []string:
		names = filters
	case []any:
		results := make([]Filter, 0, len(filters))
		for _, filter := range filters {
			result, err := normalizeFilter(filter)
			if err != nil {
				return nil, err
			}
			if result != nil {
				results = append(results, result)
			}
		}
		return results, nil
	default:
		if reflect.TypeOf(value).Kind() == reflect.Slice {
			values := reflect.ValueOf(value)
			results := make([]Filter, 0, values.Len())
			for i := 0; i < values.Len(); i++ {
				result, err := normalizeFilter(values.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				if result != nil {
					results = append(results, result)
				}
			}
			return results, nil
		}
		return nil, fmt.Errorf("filters must be a string or list")
	}

	results := make([]Filter, 0, len(names))
	for _, name := range names {
		result, err := normalizeFilter(name)
		if err != nil {
			return nil, err
		}
		if result != nil {
			results = append(results, result)
		}
	}
	return results, nil
}

func normalizeFilter(value any) (Filter, error) {
	switch filter := value.(type) {
	case string:
		name := strings.TrimSpace(filter)
		if name == "" {
			return nil, nil
		}
		return newFilter(name, nil)
	case map[string]any:
		return normalizeFilterMap(filter)
	case map[any]any:
		options := make(map[string]any, len(filter))
		for key, value := range filter {
			options[cast.ToString(key)] = value
		}
		return normalizeFilterMap(options)
	default:
		return nil, fmt.Errorf("assets filter must be a string or object")
	}
}

func normalizeFilterMap(options map[string]any) (Filter, error) {
	name := strings.TrimSpace(cast.ToString(options["name"]))
	if name == "" {
		return nil, fmt.Errorf("assets filter object requires name")
	}

	filterOptions := make(map[string]any, len(options))
	for key, value := range options {
		if key == "name" {
			continue
		}
		filterOptions[key] = value
	}
	return newFilter(name, filterOptions)
}

func newFilter(name string, options map[string]any) (Filter, error) {
	switch name {
	case assetFilterCSSMin, assetFilterJSMin:
		if len(options) > 0 {
			return nil, fmt.Errorf("assets filter %q does not accept options", name)
		}
		return &MinifyFilter{name: name}, nil
	case assetFilterImage:
		return newImageFilter(options)
	default:
		return nil, fmt.Errorf("unknown assets filter %q", name)
	}
}
