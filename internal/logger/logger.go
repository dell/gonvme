package logger

import (
	"context"
	"fmt"
	"log"
)

var logger Logger

func init() {
	logger = &DummyLogger{}
}

// SetLogger - set custom logger
func SetLogger(customLogger Logger) {
	logger = customLogger
}

// Logger logging interface for gonvme
type Logger interface {
	Info(ctx context.Context, format string, args ...interface{})
	Debug(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, format string, args ...interface{})
}

// DummyLogger - placeholder for default logger
type DummyLogger struct{}

// Info - log info using default logger
func (dl *DummyLogger) Info(ctx context.Context, format string, args ...interface{}) {
	log.Print("INFO: " + fmt.Sprintf(format, args...))
}

// Debug - log debug using default logger
func (dl *DummyLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	log.Print("DEBUG: " + fmt.Sprintf(format, args...))
}

// Error - log error using default logger
func (dl *DummyLogger) Error(ctx context.Context, format string, args ...interface{}) {
	log.Print("ERROR: " + fmt.Sprintf(format, args...))
}

// Info - log info using custom logger
func Info(ctx context.Context, format string, args ...interface{}) {
	logger.Info(ctx, format, args...)
}

// Debug - log debug using custom logger
func Debug(ctx context.Context, format string, args ...interface{}) {
	logger.Debug(ctx, format, args...)
}

// Error - log error using custom logger
func Error(ctx context.Context, format string, args ...interface{}) {
	logger.Error(ctx, format, args...)
}
