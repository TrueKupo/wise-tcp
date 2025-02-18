package log

import (
	"fmt"
	"log"
)

type stdLogger struct {
	exitHandler exitFunc
}

type exitFunc func(code int)

type logLevel string

const (
	levelInfo  logLevel = "INFO"
	levelWarn  logLevel = "WARN"
	levelDebug logLevel = "DEBUG"
	levelError logLevel = "ERROR"
	levelFatal logLevel = "FATAL"
)

func (l *stdLogger) Info(args ...interface{}) {
	l.log(levelInfo, args...)
}

func (l *stdLogger) Warn(args ...interface{}) {
	l.log(levelWarn, args...)
}

func (l *stdLogger) Error(args ...interface{}) {
	l.log(levelError, args...)
}

func (l *stdLogger) Debug(args ...interface{}) {
	l.log(levelDebug, args...)
}

func (l *stdLogger) Fatal(args ...interface{}) {
	l.log(levelFatal, args...)
	l.exitHandler(1)
}

func (l *stdLogger) Infof(format string, args ...interface{}) {
	l.logf(levelInfo, format, args...)
}

func (l *stdLogger) Warnf(msg string, args ...interface{}) {
	l.logf(levelWarn, msg, args...)
}

func (l *stdLogger) Errorf(format string, args ...interface{}) {
	l.logf(levelError, format, args...)
}

func (l *stdLogger) Debugf(format string, args ...interface{}) {
	l.logf(levelDebug, format, args...)
}

func (l *stdLogger) Fatalf(format string, args ...interface{}) {
	l.logf(levelFatal, format, args...)
	l.exitHandler(1)
}

func (l *stdLogger) log(level logLevel, args ...interface{}) {
	message := "[" + string(level) + "]"
	for _, arg := range args {
		message += fmt.Sprintf(" %v", arg)
	}
	log.Print(message)
}

func (l *stdLogger) logf(level logLevel, format string, args ...interface{}) {
	log.Printf("["+string(level)+"] "+format, args...)
}
