/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: main.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-04-02 17:57:58 (CST)
 Last Update: Saturday 2019-09-07 17:22:15 (CST)
		  By:
 Description:
 *********************************************************************************/
package main

import (
	"flag"
	"snow/core"
	"snow/reader"
	"snow/writer"
)

// VERSION ..
const VERSION = "0.1.0"

// main ..
func main() {
	var (
		version  bool
		confpath string
		server   bool
		develop  bool
		publish  bool
	)

	flag.BoolVar(&version, "v", false, "show version.")
	flag.BoolVar(&server, "s", false, "start web server.")
	flag.BoolVar(&develop, "d", false, "develop web server.")
	flag.BoolVar(&publish, "p", false, "publish web server.")
	flag.StringVar(&confpath, "c", "etc/config.toml", "load config path")
	flag.Parse()

	core.G.Conf = core.Parse(confpath)
	core.G.Server = server
	core.G.Develop = develop
	core.G.Version = VERSION
	core.Init()

	reader.Start()
	writer.Start()
	// server.Start()
}
