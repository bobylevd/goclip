package main

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"os"

	"github.com/sirupsen/logrus"
)

var log = &logrus.Logger{
	Out:       os.Stdout,
	Formatter: new(logrus.TextFormatter),
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.DebugLevel,
}

func init() {
	log.SetOutput(&lumberjack.Logger{
		Filename: "application.log",
		MaxAge:   5,
	})
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.DebugLevel)
}
