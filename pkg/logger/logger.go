package logger

import (
	"fmt"
	"os"
	"strings"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	CRITICAL
)

var levelNames = map[Level]string{
	DEBUG:    "DEBUG",
	INFO:     "INFO",
	WARNING:  "WARNING",
	ERROR:    "ERROR",
	CRITICAL: "CRITICAL",
}

var level = WARNING // Default to WARNING level

func init() {
	// Set log level from environment variable
	if levelStr := os.Getenv("JUDGMENT_LOG_LEVEL"); levelStr != "" {
		SetLevelFromString(levelStr)
	}
}

func SetLevel(l Level) {
	level = l
}

func SetLevelFromString(levelStr string) {
	switch strings.ToLower(levelStr) {
	case "debug":
		level = DEBUG
	case "info":
		level = INFO
	case "warning", "warn":
		level = WARNING
	case "error":
		level = ERROR
	case "critical":
		level = CRITICAL
	default:
		level = WARNING
	}
}

func log(l Level, format string, args ...interface{}) {
	if l < level {
		return
	}

	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s\n", levelNames[l], message)
}

func Debug(format string, args ...interface{}) {
	log(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	log(INFO, format, args...)
}

func Warning(format string, args ...interface{}) {
	log(WARNING, format, args...)
}

func Error(format string, args ...interface{}) {
	log(ERROR, format, args...)
}

func Critical(format string, args ...interface{}) {
	log(CRITICAL, format, args...)
}
