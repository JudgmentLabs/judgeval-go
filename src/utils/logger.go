package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JudgmentLabs/judgeval-go/src/env"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Gray   = "\033[90m"
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

var levelColors = map[Level]string{
	DEBUG:    Gray,
	INFO:     Gray,
	WARNING:  Yellow,
	ERROR:    Red,
	CRITICAL: Red,
}

type Logger struct {
	mu           sync.Mutex
	initialized  bool
	currentLevel Level
	useColor     bool
	output       io.Writer
}

var defaultLogger = &Logger{
	currentLevel: WARNING,
	useColor:     true,
	output:       os.Stdout,
}

func (l *Logger) initialize() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.initialized {
		return
	}

	noColor := env.JudgmentNoColor
	l.useColor = noColor == "" && isTerminal()

	logLevel := strings.ToLower(env.JudgmentLogLevel)
	l.loadLogLevel(logLevel)

	l.initialized = true
}

func (l *Logger) loadLogLevel(logLevel string) {
	switch logLevel {
	case "debug":
		l.currentLevel = DEBUG
	case "info":
		l.currentLevel = INFO
	case "warning", "warn":
		l.currentLevel = WARNING
	case "error":
		l.currentLevel = ERROR
	case "critical":
		l.currentLevel = CRITICAL
	default:
		l.currentLevel = WARNING
	}
}

func (l *Logger) setLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.currentLevel = level
}

func (l *Logger) setUseColor(useColor bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.useColor = useColor
}

func (l *Logger) setOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

func (l *Logger) log(level Level, message string) {
	l.initialize()

	if level < l.currentLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf("%s - judgeval - %s - %s",
		timestamp, levelNames[level], message)

	if l.useColor {
		formattedMessage = levelColors[level] + formattedMessage + Reset
	}

	fmt.Fprintln(l.output, formattedMessage)
}

func (l *Logger) Debug(message string) {
	l.log(DEBUG, message)
}

func (l *Logger) Info(message string) {
	l.log(INFO, message)
}

func (l *Logger) Warning(message string) {
	l.log(WARNING, message)
}

func (l *Logger) Error(message string) {
	l.log(ERROR, message)
}

func (l *Logger) Critical(message string) {
	l.log(CRITICAL, message)
}

var DefaultLogger = defaultLogger

func SetLevel(level Level) {
	defaultLogger.setLevel(level)
}

func SetUseColor(useColor bool) {
	defaultLogger.setUseColor(useColor)
}

func SetOutput(output io.Writer) {
	defaultLogger.setOutput(output)
}

func Debug(message string) {
	defaultLogger.Debug(message)
}

func Info(message string) {
	defaultLogger.Info(message)
}

func Warning(message string) {
	defaultLogger.Warning(message)
}

func Error(message string) {
	defaultLogger.Error(message)
}

func Critical(message string) {
	defaultLogger.Critical(message)
}

func isTerminal() bool {

	return isatty(os.Stdout.Fd())
}

func isatty(fd uintptr) bool {

	return true
}
