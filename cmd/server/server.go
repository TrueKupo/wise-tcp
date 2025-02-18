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
	"wise-tcp/pkg/log"
	"wise-tcp/pkg/zap"
)

type Config struct {
	App    AppConfig       `yaml:"app"`
	Server server.Config   `yaml:"server"`
	Guard  pow.GuardConfig `yaml:"guard"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Prod bool   `yaml:"isProd"`
}

func main() {
	cfg := mustLoadConfig()

	log.Info("Server application starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initLogger(cfg.App)

	guard, err := initServerGuard(cfg.Guard)
	if err != nil {
		log.Fatalf("Failed to initialize server guard: %v", err)
	}

	qh, err := handler.NewQuote()
	if err != nil {
		log.Fatalf("Failed to initialize quote handler: %v", err)
	}

	srv, err := server.NewServer(
		server.WithConfig(cfg.Server),
		server.WithGuard(guard),
		server.WithHandler(qh),
	)
	if err != nil {
		log.Fatalf("Failed to initialize tcp server: %v", err)
	}

	go func() {
		if err = srv.Start(ctx); err != nil {
			log.Fatalf("Server start failed: %v", err)
		}
	}()

	gracefulManager := initGraceful(srv)
	if err = gracefulManager.Start(ctx); err != nil {
		log.Errorf("Shutdown failed: %v", err)
	}

	log.Infof("Application stopped")
}

func mustLoadConfig() *Config {
	return config.MustLoad[Config]("cfg/server.yml",
		config.WithEnvMapper[Config](applyConfigMapping))
}

func initLogger(cfg AppConfig) {
	logger, err := zap.New(
		zap.WithName(cfg.Name),
		zap.WithProd(cfg.Prod),
	)
	if err != nil {
		log.Errorf("Failed to initialize zap logger: %v", err)
		return
	}

	log.SetLogger(logger)
}

func initServerGuard(cfg pow.GuardConfig) (server.Guard, error) {
	var difficulty int
	if cfg.PowDifficulty == 0 {
		difficulty = 20
	} else if cfg.PowDifficulty < 0 {
		return nil, errors.New("difficulty must be positive or 0")
	} else {
		difficulty = cfg.PowDifficulty
	}
	provider := hashcash.NewProvider(hashcash.WithDifficulty(difficulty))

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
