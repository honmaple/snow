package builder

import (
	"errors"
	"fmt"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/config"
)

const (
	DAY90 = 90 * 24 * time.Hour
	DAY10 = 10 * 24 * time.Hour
	DAY7  = 7 * 24 * time.Hour
	DAY   = 24 * time.Hour
)

type Filter struct {
	conf *config.Config
}

func (f *Filter) timeSince(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	v, ok := in.Interface().(time.Time)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:date",
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

func (f *Filter) groupBy(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	pages, ok := in.Interface().(page.Pages)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:groupBy",
			OrigError: errors.New("filter input argument must be of type 'Pages'"),
		}
	}
	key := ""
	if param != nil {
		key = param.String()
	}
	sections := pages.GroupBy(key)
	if sections == nil {
		return nil, &pongo2.Error{
			Sender:    "filter:groupBy",
			OrigError: errors.New("unknown group key"),
		}
	}
	return pongo2.AsValue(sections), nil
}

func (f *Filter) absURL(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	return nil, nil
}

func (f *Filter) relURL(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	return nil, nil
}

func registerFilter(conf *config.Config) {
	f := &Filter{conf: conf}
	pongo2.RegisterFilter("absURL", f.absURL)
	pongo2.RegisterFilter("relURL", f.relURL)
	pongo2.RegisterFilter("groupBy", f.groupBy)
	pongo2.RegisterFilter("timesince", f.timeSince)
}
