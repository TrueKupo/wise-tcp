package log

import (
	std "log"
	"os"
)

type exitFunc func(code int)

type defaultLogger struct {
	exitHandler exitFunc
}

var DefaultLogger = &defaultLogger{
	exitHandler: os.Exit,
}

func Default() Logger {
	return DefaultLogger
}

func (l *defaultLogger) Info(msg string, args ...interface{}) {
	std.Printf("[INFO] "+msg, args...)
}

func (l *defaultLogger) Warn(msg string, args ...interface{}) {
	std.Printf("[WARN] "+msg, args...)
}

func (l *defaultLogger) Error(msg string, args ...interface{}) {
	std.Printf("[ERROR] "+msg, args...)
}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	std.Printf("[DEBUG] "+msg, args...)
}

func (l *defaultLogger) Fatal(msg string, args ...interface{}) {
	std.Printf("[FATAL] "+msg, args...)
	l.exitHandler(1)
}

func (l *defaultLogger) Flush() error {
	return nil
}
