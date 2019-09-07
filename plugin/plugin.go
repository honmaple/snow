/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: plugin.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-09-05 19:49:05 (CST)
 Last Update: Saturday 2019-09-07 17:21:50 (CST)
		  By:
 Description:
 *********************************************************************************/
package plugin

import (
	"snow/core"
	"snow/reader"
	"snow/writer"
)

type PluginType interface {
	Register()
}

type Plugin struct {
	Name string
}

var plugins = []PluginType{NewSortPlugin()}

func (s *Plugin) AddVar(key string, value interface{}) {
	core.G.AddVar(key, value)
}

func (s *Plugin) AddReader(r reader.ReaderType) {
	reader.Add(r)
}

func (s *Plugin) AddWriter(w writer.WriterType) {
	writer.Add(w)
}

func Add(plugin PluginType) {
	plugins = append(plugins, plugin)
}

func Clean() {
	plugins = make([]PluginType, 0)
}

func Init() {
	for _, plugin := range plugins {
		plugin.Register()
	}
}
