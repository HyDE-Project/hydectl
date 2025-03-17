// filepath: /home/khing/The HyDE Project/hydectl/internal/logging/logger.go
package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var logger = log.New(os.Stdout)

func SetupLogging() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "silent" // Default log level to silent (no logs)
	}

	var parsedLevel log.Level
	var err error
	if level == "silent" {
		parsedLevel = log.FatalLevel + 1 // Set to a level higher than fatal to silence logs
	} else {
		parsedLevel, err = log.ParseLevel(level)
		if err != nil {
			logger.Errorf("Invalid log level: %s", level)
			parsedLevel = log.InfoLevel // Fallback to info level
		}
	}

	logger.SetLevel(parsedLevel)
	logger.Debugf("Log level set to %s", parsedLevel)
}

func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}
