package zap

import (
	"go.uber.org/zap"
)

type Config struct {
	Prod bool
	Name string
}

type Logger struct {
	logger *zap.SugaredLogger
	isProd bool
	name   string
}

type Option func(l *Logger)

func WithProd(v bool) Option {
	return func(l *Logger) {
		l.isProd = v
	}
}

func WithName(name string) Option {
	return func(l *Logger) {
		l.name = name
	}
}

func New(opts ...Option) (*Logger, error) {
	var zapLogger *zap.Logger
	var err error

	l := &Logger{}

	for _, opt := range opts {
		opt(l)
	}

	if l.isProd {
		zapLogger, err = zap.NewProduction()
	} else {
		zapLogger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	options := []zap.Option{
		zap.AddCallerSkip(2),
		zap.AddCaller(),
	}
	if l.name != "" {
		options = append(options, zap.Fields(zap.String("name", l.name)))
	}

	zapLogger = zapLogger.WithOptions(options...)

	l.logger = zapLogger.Sugar()

	return l, nil
}

func (z *Logger) Info(args ...interface{}) {
	z.logger.Info(args)
}

func (z *Logger) Warn(args ...interface{}) {
	z.logger.Warn(args)
}

func (z *Logger) Error(args ...interface{}) {
	z.logger.Error(args)
}

func (z *Logger) Debug(args ...interface{}) {
	z.logger.Debug(args)
}

func (z *Logger) Fatal(args ...interface{}) {
	z.logger.Fatal(args)
}

func (z *Logger) Infof(format string, args ...interface{}) {
	z.logger.Infof(format, args...)
}

func (z *Logger) Warnf(format string, args ...interface{}) {
	z.logger.Warnf(format, args...)
}

func (z *Logger) Errorf(format string, args ...interface{}) {
	z.logger.Errorf(format, args...)
}

func (z *Logger) Debugf(format string, args ...interface{}) {
	z.logger.Debugf(format, args...)
}

func (z *Logger) Fatalf(format string, args ...interface{}) {
	z.logger.Fatalf(format, args...)
}
