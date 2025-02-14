package graceful

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"wise-tcp/pkg/log"
)

const defaultTimeout = 5 * time.Second

type manager struct {
	services []Service
	timeout  time.Duration
	log      log.Logger
}

func (m *manager) Register(svc Service) {
	m.services = append(m.services, svc)
}

type Option func(*manager)

func WithTimeout(timeout time.Duration) Option {
	return func(m *manager) {
		m.timeout = timeout
	}
}

func WithLogger(logger log.Logger) Option {
	return func(m *manager) {
		m.log = logger
	}
}

func NewManager(opts ...Option) Manager {
	m := &manager{
		services: make([]Service, 0),
		timeout:  defaultTimeout,
		log:      log.Default(),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *manager) Start(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	select {
	case sig := <-sigChan:
		m.log.Info("Received signal: %v", sig)
	case <-ctx.Done():
		m.log.Info("Shutdown externally triggered")
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	return m.Stop(ctx)
}

func (m *manager) Stop(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(m.services))

	for _, svc := range m.services {
		wg.Add(1)
		go func(s Stopper) {
			defer wg.Done()
			if err := s.Stop(ctx); err != nil {
				errCh <- fmt.Errorf("service %s failed to stop: %w", s, err)
				m.log.Error("Service %s failed to stop: %v", s, err)
			} else {
				m.log.Info("Service %s stopped successfully", s)
			}
		}(svc)
	}

	wg.Wait()
	close(errCh)

	errs := make([]error, 0, len(m.services))
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred during shutdown:\n%v", errs)
	}

	return nil
}
