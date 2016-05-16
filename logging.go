package prometheusproxy

import (
	log "github.com/Sirupsen/logrus"
)

var Logger *log.Entry

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: false})
	log.SetLevel(log.DebugLevel)
	Logger = log.WithFields(log.Fields{"app": "prometheusproxy"})
}

func configLogger(logLevel string) bool {
	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		return true
	case "info":
		log.SetLevel(log.InfoLevel)
		return true
	case "error":
		log.SetLevel(log.ErrorLevel)
		return true
	}
	return false
}
