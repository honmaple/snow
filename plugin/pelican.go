/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: pelican.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-09-05 19:49:38 (CST)
 Last Update: Friday 2019-09-06 00:55:04 (CST)
		  By:
 Description:
 *********************************************************************************/
package plugin

import (
	"snow/core"
)

type PelicanPlugin struct {
	Plugin
}

func NewPelicanPlugin() *PelicanPlugin {
	return &PelicanPlugin{}
}

func (s *PelicanPlugin) Register() {
	conf := core.G.Conf
	s.AddVar("SITEURL", conf.Site.URL)
	s.AddVar("SITENAME", conf.Site.Title)
	s.AddVar("DEFAULT_LANG", conf.Lang)
}
