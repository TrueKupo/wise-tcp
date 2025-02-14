package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"wise-tcp/internal/graceful"
	"wise-tcp/internal/handler"
	"wise-tcp/internal/pow"
	"wise-tcp/internal/pow/hashcash"
	"wise-tcp/internal/server"
	"wise-tcp/pkg/config"
	"wise-tcp/pkg/factory"
	"wise-tcp/pkg/log"
	"wise-tcp/pkg/log/zap"
)

type Config struct {
	Logger log.Config      `yaml:"logger"`
	Server server.Config   `yaml:"server"`
	Guard  pow.GuardConfig `yaml:"guard"`
}

func main() {
	cfg := mustLoadConfig()

	logger := initLogger(cfg.Logger)
	logger.Info("Server application starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceFactory := factory.New(
		factory.WithLogger(logger),
	)

	guard, err := initServerGuard(serviceFactory, cfg.Guard)
	if err != nil {
		logger.Fatal("Failed to initialize server guard: %v", err)
	}

	qh, err := handler.NewQuote(serviceFactory)
	if err != nil {
		logger.Fatal("Failed to initialize quote handler: %v", err)
	}

	srv, err := server.NewServer(
		serviceFactory,
		server.WithConfig(cfg.Server),
		server.WithGuard(guard),
		server.WithHandler(qh),
	)
	if err != nil {
		logger.Fatal("Failed to initialize tcp server: %v", err)
	}

	go func() {
		if err = srv.Start(ctx); err != nil {
			logger.Fatal("Server start failed: %v", err)
		}
	}()

	gracefulManager := initGraceful(srv)
	if err = gracefulManager.Start(ctx); err != nil {
		logger.Error("Shutdown failed: %v", err)
	}

	logger.Info("Application stopped")
}

func mustLoadConfig() *Config {
	return config.MustLoad[Config]("cfg/server.yml",
		config.WithEnvMapper[Config](applyConfigMapping))
}

func initLogger(cfg log.Config) log.Logger {
	return zap.Init(cfg)
}

func initServerGuard(fc factory.Factory, cfg pow.GuardConfig) (server.Guard, error) {
	var difficulty int
	if cfg.PowDifficulty == 0 {
		difficulty = 20
	} else if cfg.PowDifficulty < 0 {
		return nil, errors.New("difficulty must be positive or 0")
	} else {
		difficulty = cfg.PowDifficulty
	}
	provider := hashcash.NewProvider(fc, hashcash.WithDifficulty(difficulty))

	return pow.NewGuard(provider), nil
}

func initGraceful(services ...graceful.Service) graceful.Manager {
	manager := graceful.NewManager(
		graceful.WithTimeout(5 * time.Second),
	)
	for _, svc := range services {
		manager.Register(svc)
	}
	return manager
}

func applyConfigMapping(v *viper.Viper) error {
	if err := v.BindEnv("server.port", "PORT"); err != nil {
		return fmt.Errorf("failed to bind PORT: %w", err)
	}
	if err := v.BindEnv("server.maxConn", "MAX_CONN"); err != nil {
		return fmt.Errorf("failed to bind MAX_CONN: %w", err)
	}
	if err := v.BindEnv("guard.powDifficulty", "POW_DIFFICULTY"); err != nil {
		return fmt.Errorf("failed to bind POW_DIFFICULTY: %w", err)
	}

	return nil
}
