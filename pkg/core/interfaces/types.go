package interfaces

import "wise-tcp/pkg/log"

type LoggerAware interface {
	SetLogger(logger log.Logger)
}
