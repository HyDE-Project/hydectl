// filepath: /home/khing/The HyDE Project/hydectl/internal/logging/logging.go
package logging

import (
	"log"
	"os"
)

var (
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logLevel    = "silent"
)

func SetupLogging() {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		logLevel = level
	}
}

func Debugf(format string, v ...interface{}) {
	if logLevel == "debug" {
		debugLogger.Printf(format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if logLevel == "info" || logLevel == "debug" {
		infoLogger.Printf(format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if logLevel != "silent" {
		errorLogger.Printf(format, v...)
	}
}
