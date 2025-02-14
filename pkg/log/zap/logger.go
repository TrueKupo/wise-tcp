package zap

import (
	"go.uber.org/zap"
	"wise-tcp/pkg/log"
)

type Logger struct {
	logger *zap.SugaredLogger
	raw    *zap.Logger
}

func New(cfg log.Config) (log.Logger, error) {
	var zapLogger *zap.Logger
	var err error

	if cfg.Prod {
		zapLogger, err = zap.NewProduction()
	} else {
		zapLogger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	options := []zap.Option{
		zap.AddCallerSkip(1),
		zap.AddCaller(),
	}

	if cfg.Name != "" {
		options = append(options, zap.Fields(zap.String("name", cfg.Name)))
	}

	zapLogger = zapLogger.WithOptions(options...)
	return &Logger{
		logger: zapLogger.Sugar(),
		raw:    zapLogger,
	}, nil
}

func Init(cfg log.Config) log.Logger {
	logger, err := New(cfg)
	if err != nil {
		return fallbackLogger(err, cfg)
	}
	return logger
}

func (z *Logger) Info(msg string, args ...interface{}) {
	z.logger.Infof(msg, args...)
}

func (z *Logger) Warn(msg string, args ...interface{}) {
	z.logger.Warnf(msg, args...)
}

func (z *Logger) Error(msg string, args ...interface{}) {
	z.logger.Errorf(msg, args...)
}

func (z *Logger) Debug(msg string, args ...interface{}) {
	z.logger.Debugf(msg, args...)
}

func (z *Logger) Fatal(msg string, args ...interface{}) {
	z.logger.Fatalf(msg, args...)
}

func (z *Logger) Flush() error {
	return z.raw.Sync()
}

func fallbackLogger(err error, cfg log.Config) log.Logger {
	fallback := log.Default()
	fallback.Error("Failed to initialize zap logger (cfg: %#v): %v", cfg, err)
	return fallback
}
