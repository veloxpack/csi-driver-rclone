// Package logger defines the Cloudinary Logger
package logger

import (
	"log"
)

// Level represents the level of the logger.
type Level int8

// NONE level of the logger.
const NONE Level = 0

// ERROR level of the logger.
const ERROR Level = 1

// DEBUG level of the logger.
const DEBUG Level = 2

// LogWriter is an interface for a log writer.
type LogWriter interface {
	Debug(v ...interface{})
	Error(v ...interface{})
}

// GoLog is a struct for go log writer.
type GoLog struct{}

// Debug prints debug messages.
func (g *GoLog) Debug(v ...interface{}) {
	log.Println("cloudinary debug", v)
}

// Error prints error messages.
func (g *GoLog) Error(v ...interface{}) {
	log.Println("cloudinary error", v)
}

// Logger is the logger struct.
type Logger struct {
	Writer LogWriter
	level  Level
}

// SetLevel sets the logger level.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// Debug writes error messages.
func (l *Logger) Error(v ...interface{}) {
	if l.level >= ERROR {
		l.Writer.Error(v...)
	}
}

// Debug writes debug messages.
func (l *Logger) Debug(v ...interface{}) {
	if l.level == DEBUG {
		l.Writer.Debug(v...)
	}
}

// New returns a new logger instance.
func New() *Logger {
	return &Logger{
		Writer: &GoLog{},
		level:  ERROR,
	}
}
