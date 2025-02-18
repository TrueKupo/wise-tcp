package server

import (
	"context"
	"errors"
	"net"
	"time"

	"golang.org/x/sync/semaphore"
)

type ThrottleConfig struct {
	MaxConn int           `mapstructure:"max" env:"MAX_CONN"`
	Policy  string        `mapstructure:"policy"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type ThrottlePolicy string

const (
	BlockPolicy  ThrottlePolicy = "block"
	RejectPolicy ThrottlePolicy = "reject"
	DropPolicy   ThrottlePolicy = "drop"
)

var (
	ErrConnRejected = errors.New("connection rejected: too many connections")
	ErrConnDropped  = errors.New("connection dropped: too many connections")
)

type Throttle struct {
	maxConn int64
	sem     *semaphore.Weighted
	policy  ThrottlePolicy
	timeout time.Duration
}

func NewThrottle(cfg ThrottleConfig) *Throttle {
	return &Throttle{
		sem:     semaphore.NewWeighted(int64(cfg.MaxConn)),
		maxConn: int64(cfg.MaxConn),
		policy:  ThrottlePolicy(cfg.Policy),
		timeout: cfg.Timeout,
	}
}

func (t *Throttle) Acquire(ctx context.Context, conn net.Conn) error {
	switch t.policy {
	case BlockPolicy:
		return t.sem.Acquire(ctx, 1)

	case RejectPolicy:
		rejectCtx, cancel := context.WithTimeout(ctx, t.timeout)
		defer cancel()

		err := t.sem.Acquire(rejectCtx, 1)
		if err != nil {
			_, _ = conn.Write([]byte("Service Unavailable\n"))
			_ = conn.Close()
			return ErrConnRejected
		}
		return nil

	case DropPolicy:
		if t.sem.TryAcquire(1) {
			return nil
		}

		_ = conn.Close()
		return ErrConnDropped

	default:
		return errors.New("unrecognized ThrottlePolicy")
	}
}

func (t *Throttle) Release() {
	t.sem.Release(1)
}
