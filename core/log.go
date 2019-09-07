/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: log.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-04-05 00:18:35 (CST)
 Last Update: Thursday 2019-09-05 00:52:52 (CST)
		  By:
 Description:
 *********************************************************************************/
package core

import (
	"github.com/sirupsen/logrus"
	"os"
)

// Log ..
type Log struct {
	Level     string
	Timestamp bool
}

// InitLogger ..
func Init() {
	level := logrus.ErrorLevel
	switch G.Conf.Log.Level {
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warn":
		level = logrus.WarnLevel
	default:
		level = logrus.ErrorLevel
	}

	G.Logger = &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			FullTimestamp: G.Conf.Log.Timestamp,
		},
		Level: level,
	}
}
