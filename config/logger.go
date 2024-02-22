package config

import (
	"io"
	"log"
	"os"
)

// Logger represents a logger with different log levels.
type Logger struct {
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	err     *log.Logger
	writer  io.Writer
}

// NewLogger creates a new logger instance with the provided prefix.
func NewLogger(p string) *Logger {
	writer := io.Writer(os.Stdout)
	logger := log.New(writer, p, log.Ldate|log.Ltime)

	return &Logger{
		err:     log.New(writer, "> ERROR ", logger.Flags()),
		info:    log.New(writer, "> INFO ", logger.Flags()),
		warning: log.New(writer, "> WARNING ", logger.Flags()),
		debug:   log.New(writer, "> DEBUG ", logger.Flags()),
		writer:  writer,
	}
}

// Debug prints debug logs.
func (l *Logger) Debug(v ...interface{}) {
	l.debug.Println(v...)
}

// Info prints info logs.
func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

// Warn prints warning logs.
func (l *Logger) Warn(v ...interface{}) {
	l.warning.Println(v...)
}

// Error prints error logs.
func (l *Logger) Error(v ...interface{}) {
	l.err.Println(v...)
}

// Debugf prints formatted debug logs.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debug.Printf(format, v...)
}

// Infof prints formatted info logs.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

// Warnf prints formatted warning logs.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.warning.Printf(format, v...)
}

// Errorf prints formatted error logs.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.err.Printf(format, v...)
}
