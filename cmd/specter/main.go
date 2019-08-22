package main

import (
	"github.com/namsral/flag"
	"github.com/newshipt/specter/internal/webServer"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var logLevel string

func init() {
	flag.StringVar(&logLevel, "loglvl", "Warn", "Log level")
}

func main() {
	flag.Parse()
	switch {
	case logLevel == "Debug":
		log.SetLevel(logrus.DebugLevel)
	case logLevel == "Info":
		log.SetLevel(logrus.InfoLevel)
	case logLevel == "Warn":
		log.SetLevel(logrus.WarnLevel)
	case logLevel == "Error":
		log.SetLevel(logrus.ErrorLevel)
	case logLevel == "Fatal":
		log.SetLevel(logrus.FatalLevel)
	}

	webServer.Start()
}
