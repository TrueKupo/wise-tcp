package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"wise-tcp/internal/auth"
	"wise-tcp/pkg/core/build"
	"wise-tcp/pkg/log"
)

type Config struct {
	Port     int            `mapstructure:"port" env:"PORT"`
	Timeout  time.Duration  `mapstructure:"timeout"`
	Throttle ThrottleConfig `mapstructure:"throttle" env:"MAX_CONN"`
}

func (c Config) Name() string {
	return "tcp-server"
}

type RequestHandler interface {
	Handle(ctx context.Context, rw io.ReadWriter) error
}

type TCPServer struct {
	addr     string
	listener net.Listener
	cfg      Config
	handler  *connHandler
	wg       sync.WaitGroup
}

type Option func(*TCPServer)

func Builder(cfg Config) build.Builder {
	return func(i *build.Injector) (any, error) {
		h, err := build.Extract[RequestHandler](i, "server.handler")
		if err != nil {
			return nil, err
		}

		a, err := build.Extract[auth.RequestAuthorizer](i, "server.auth")
		if err != nil {
			log.Error(err)
		}

		return &TCPServer{
			cfg:  cfg,
			addr: fmt.Sprintf(":%d", cfg.Port),
			handler: &connHandler{
				throttle:   NewThrottle(cfg.Throttle),
				auth:       a,
				reqHandler: h,
			},
		}, nil
	}
}

func (s *TCPServer) Start(ctx context.Context) error {
	if s.listener != nil {
		return fmt.Errorf("server is already running")
	}

	log.Debugf("Initializing server with config: %#v", s.cfg)

	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	log.Infof("TCP server listening on %s", s.addr)

	go s.acceptLoop(ctx)

	return nil
}

func (s *TCPServer) acceptLoop(ctx context.Context) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Info("Server shutting down due to context cancellation")
				return
			default:
				if errors.Is(err, net.ErrClosed) {
					// listener closed, ignore
					return
				}
				log.Error("Failed to accept connection: %v", err)
			}
			continue
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			cctx, cancel := context.WithDeadline(ctx, time.Now().Add(s.cfg.Timeout))
			defer cancel()
			s.handler.Handle(cctx, conn)
		}()
	}
}

func (s *TCPServer) Stop(ctx context.Context) error {
	log.Info("Shutting down TCP server...")

	if s.listener != nil {
		if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("All connections closed")
	case <-ctx.Done():
		log.Warn("Shutdown timeout exceeded, forcing shutdown")
	}

	return nil
}

func (s *TCPServer) String() string {
	return fmt.Sprintf("TCPServer on %s", s.addr)
}
