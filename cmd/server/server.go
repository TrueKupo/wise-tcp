package main

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"wise-tcp/internal/handler"
	"wise-tcp/internal/pow"
	"wise-tcp/internal/server"
	"wise-tcp/pkg/config"
	"wise-tcp/pkg/core"
	"wise-tcp/pkg/log"
	"wise-tcp/pkg/zap"
)

type Config struct {
	App    AppConfig     `yaml:"app"`
	Server server.Config `yaml:"server"`
	//Guard  pow.GuardConfig `yaml:"guard"`
	Pow pow.Config `yaml:"pow"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Prod bool   `yaml:"isProd"`
}

func main() {
	cfg := mustLoadConfig()
	initLogger(cfg.App)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Initializing application...")

	app := core.NewApp()
	err := app.BuildUnits(
		core.UnitBuilder{Builder: pow.AuthBuilder(cfg.Pow), Name: "server.auth"},
		core.UnitBuilder{Builder: handler.Builder(), Name: "server.handler"},
		core.UnitBuilder{Builder: server.Builder(cfg.Server), Name: "server"},
	)
	if err != nil {
		log.Fatalf("Failed to build app: %v", err)
	}

	if err = app.Go(ctx); err != nil {
		log.Fatal(err)
	}

	log.Infof("Application finished with state %s", app.State())
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

func applyConfigMapping(v *viper.Viper) error {
	if err := v.BindEnv("server.port", "PORT"); err != nil {
		return fmt.Errorf("failed to bind PORT: %w", err)
	}
	if err := v.BindEnv("server.throttle.max", "MAX_CONN"); err != nil {
		return fmt.Errorf("failed to bind MAX_CONN: %w", err)
	}
	if err := v.BindEnv("pow.diff", "POW_DIFFICULTY"); err != nil {
		return fmt.Errorf("failed to bind POW_DIFFICULTY: %w", err)
	}
	if err := v.BindEnv("pow.async", "POW_ASYNC"); err != nil {
		return fmt.Errorf("failed to bind POW_ASYNC: %w", err)
	}
	if err := v.BindEnv("pow.redis", "REDIS_ADDR"); err != nil {
		return fmt.Errorf("failed to bind REDIS_ADDR: %w", err)
	}

	return nil
}
