package logger

import (
	"log"
	"strings"

	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/env"
)

const (
	ERROR LogLevel = iota
	WARN
	INFO
	DEBUG
)

type LogLevel int

type Logger interface {
	Error(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Info(format string, v ...interface{})
	Debug(format string, v ...interface{})
}

type LocalLogger struct {
	logLevel LogLevel
}

func New() *LocalLogger {
	return &LocalLogger{
		logLevel: setLogLevel(),
	}
}

func setLogLevel() LogLevel {
	logLevelStr := strings.ToLower(env.Get("ZSH_AI_SUGGESTIONS_LOG_LEVEL", "info"))
	switch logLevelStr {
	case "error":
		return ERROR
	case "warn":
		return WARN
	case "info":
		return INFO
	case "debug":
		return DEBUG
	case "none", "off":
		return -1
	default:
		return INFO
	}
}

func (l *LocalLogger) Error(format string, v ...any) {
	if l.logLevel >= ERROR {
		log.Printf("[ERROR] "+format, v...)
	}
}

func (l *LocalLogger) Warn(format string, v ...any) {
	if l.logLevel >= WARN {
		log.Printf("[WARN] "+format, v...)
	}
}

func (l *LocalLogger) Info(format string, v ...any) {
	if l.logLevel >= INFO {
		log.Printf("[INFO] "+format, v...)
	}
}

func (l *LocalLogger) Debug(format string, v ...any) {
	if l.logLevel >= DEBUG {
		log.Printf("[DEBUG] "+format, v...)
	}
}
