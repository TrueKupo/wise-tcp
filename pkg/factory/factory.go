package factory

import (
	"wise-tcp/pkg/log"
)

type Factory interface {
	Logger() log.Logger
}

type factory struct {
	logger log.Logger
}

type Option func(*factory)

func WithLogger(logger log.Logger) Option {
	return func(f *factory) {
		f.logger = logger
	}
}

func New(opts ...Option) Factory {
	f := &factory{}

	for _, opt := range opts {
		opt(f)
	}

	if f.logger == nil {
		f.logger = log.Default()
	}

	return f
}

func (f *factory) Logger() log.Logger {
	return f.logger
}
