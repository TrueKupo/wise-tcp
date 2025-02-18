package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"wise-tcp/internal/graceful"
	"wise-tcp/pkg/log"
)

type Config struct {
	Port    int           `yaml:"port" env:"PORT"`
	Timeout time.Duration `yaml:"timeout"`
	MaxConn int           `yaml:"maxConn" env:"MAX_CONN"`
}

type Server interface {
	Start(ctx context.Context) error
	graceful.Service
}

type Guard interface {
	Verify(ctx context.Context, conn net.Conn) error
}

type RequestHandler interface {
	Handle(ctx context.Context, conn net.Conn) error
}

type TCPServer struct {
	addr       string
	listener   net.Listener
	cfg        Config
	guard      Guard
	handler    RequestHandler
	activeConn int
	connLock   sync.Mutex
	wg         sync.WaitGroup
}

type Option func(*TCPServer)

func WithConfig(cfg Config) Option {
	return func(s *TCPServer) {
		s.cfg = cfg
	}
}

func WithGuard(g Guard) Option {
	return func(s *TCPServer) {
		s.guard = g
	}
}

func WithHandler(h RequestHandler) Option {
	return func(s *TCPServer) {
		s.handler = h
	}
}

var defaultConfig = Config{
	Port:    8080,
	Timeout: 5 * time.Second,
	MaxConn: 1000,
}

func NewServer(opts ...Option) (*TCPServer, error) {
	s := &TCPServer{
		cfg: defaultConfig,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.addr = fmt.Sprintf(":%d", s.cfg.Port)

	return s, nil
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

	return s.acceptConnections(ctx)
}

func (s *TCPServer) acceptConnections(ctx context.Context) error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Info("Server shutting down due to context cancellation")
				return nil
			default:
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				log.Error("Failed to accept connection: %v", err)
			}
			continue
		}

		if !s.incrementConnCount(conn) {
			continue
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(ctx, conn)
		}()
	}
}

func (s *TCPServer) incrementConnCount(conn net.Conn) bool {
	s.connLock.Lock()
	defer s.connLock.Unlock()

	if s.activeConn >= s.cfg.MaxConn {
		_, _ = conn.Write([]byte("Service currently unavailable, retry later"))
		_ = conn.Close()
		log.Warn("Connection rejected: max connections limit reached")
		return false
	}

	s.activeConn++
	return true
}

func (s *TCPServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		s.connLock.Lock()
		s.activeConn--
		s.connLock.Unlock()

		_ = conn.Close()
	}()

	if s.cfg.Timeout > 0 {
		if err := conn.SetDeadline(time.Now().Add(s.cfg.Timeout)); err != nil {
			log.Errorf("Failed to set connection deadline: %v", err)
			return
		}
	}

	if s.guard != nil {
		if err := s.guard.Verify(ctx, conn); err != nil {
			log.Warnf("Connection verification failed: %v", err)
			return
		}
	}

	log.Debugf("Connection verified successfully (addr: %s)...", conn.RemoteAddr())

	if err := s.handler.Handle(ctx, conn); err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			log.Warn("Connection timed out during processing")
		} else {
			log.Errorf("Handler failed to process request: %v", err)
		}
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
