package lib

import (
	"fmt"
)

type NullLogger struct{}

func (n NullLogger) Printf(_ string, _ ...any) {}

type Logger interface {
	Printf(string, ...any)
}

type LeveledLogger interface {
	Error(string, ...any)
	Warn(string, ...any)
	Info(string, ...any)
	Debug(string, ...any)
	Trace(string, ...any)
}

type loggerWrapper struct {
	Logger
}

func NewLeveledLogger(logger Logger) LeveledLogger {
	return &loggerWrapper{logger}
}

func (l *loggerWrapper) Error(format string, args ...any) {
	l.logWithLevel("ERROR", format, args...)
}

func (l *loggerWrapper) Warn(format string, args ...any) {
	l.logWithLevel("WARN", format, args...)
}

func (l *loggerWrapper) Info(format string, args ...any) {
	l.logWithLevel("INFO", format, args...)
}

func (l *loggerWrapper) Debug(format string, args ...any) {
	l.logWithLevel("DEBUG", format, args...)
}

func (l *loggerWrapper) Trace(format string, args ...any) {
	l.logWithLevel("TRACE", format, args...)
}

func (l *loggerWrapper) logWithLevel(level string, format string, args ...any) {
	if levelLogger, ok := l.Logger.(LeveledLogger); ok {
		switch level {
		case "ERROR":
			levelLogger.Error(format, args...)
		case "WARN":
			levelLogger.Warn(format, args...)
		case "INFO":
			levelLogger.Info(format, args...)
		case "DEBUG":
			levelLogger.Debug(format, args...)
		case "TRACE":
			levelLogger.Trace(format, args...)
		default:
			l.log(level, format, args...)
		}
		return
	}

	l.log(level, format, args...)
}

func (l *loggerWrapper) log(level string, format string, args ...any) {
	format = fmt.Sprintf("[%v] %v", level, format)
	l.Printf(format, args...)
}
