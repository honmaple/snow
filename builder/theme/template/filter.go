package template

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/utils"
)

const (
	DAY90 = 90 * 24 * time.Hour
	DAY10 = 10 * 24 * time.Hour
	DAY7  = 7 * 24 * time.Hour
	DAY   = 24 * time.Hour
)

func (t *template) timeSince(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	v, ok := in.Interface().(time.Time)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:timesince",
			OrigError: errors.New("filter input argument must be of type 'time.Time'"),
		}
	}
	if param != nil {
		if i := param.String(); i != "" {
			return pongo2.AsValue(v.Format(i)), nil
		}
	}

	since := time.Since(v)
	value := pongo2.AsValue("刚刚")
	switch {
	case since > DAY90:
		value = pongo2.AsValue(v.Format("2006年01月02日"))
	case since > DAY10:
		value = pongo2.AsValue(v.Format("01月02日"))
	case since > DAY:
		value = pongo2.AsValue(fmt.Sprintf("%d天前", since/DAY))
	case since > time.Hour:
		value = pongo2.AsValue(fmt.Sprintf("%d小时前", since/time.Hour))
	case since > time.Minute:
		value = pongo2.AsValue(fmt.Sprintf("%d分钟前", since/time.Minute))
	}
	return value, nil
}

func (t *template) absURL(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	out, err = t.relURL(in, param)
	if err != nil {
		return
	}
	v := out.Interface().(string)
	return pongo2.AsValue(t.conf.GetURL(v)), nil
}

func (t *template) relURL(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	v, ok := in.Interface().(string)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:relURL",
			OrigError: errors.New("filter input argument must be of type 'string'"),
		}
	}
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return pongo2.AsValue(v), nil
	}
	key, ok := "", false
	if param.Interface() != nil {
		key, ok = param.Interface().(string)
		if !ok {
			return nil, &pongo2.Error{
				Sender:    "filter:relURL",
				OrigError: errors.New("filter input argument must be of type 'string'"),
			}
		}
	}
	if key != "" {
		vars := map[string]string{
			"{slug}":       v,
			"{number}":     "",
			"{number:one}": "1",
		}
		v = utils.StringReplace(t.conf.GetString(key), vars)
	}
	if strings.HasPrefix(v, "/") {
		return pongo2.AsValue(v), nil
	}
	return pongo2.AsValue(utils.StringConcat("/", v)), nil
}

func (t *template) getSection(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	out, err = t.relURL(in, param)
	if err != nil {
		return
	}
	v := out.Interface().(string)
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return pongo2.AsValue(v), nil
	}
	return pongo2.AsValue(utils.StringConcat(t.conf.GetString("site.url"), v)), nil
}
