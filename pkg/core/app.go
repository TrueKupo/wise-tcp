package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wise-tcp/pkg/core/build"
	"wise-tcp/pkg/log"
)

type App struct {
	main    *Module
	factory *build.Factory
}

func NewApp() *App {
	return &App{
		main:    NewModule("main"),
		factory: build.NewFactory(),
	}
}

func (a *App) Provide(name string, item any) {
	a.factory.Injector().Register(item, name)
}

func (a *App) AddModule(module *Module) *App {
	a.main.AddModule(module)
	return a
}

func (a *App) GetModule(name string) *Module {
	return a.main.GetModule(name)
}

type UnitBuilder struct {
	Name    string
	Builder build.Builder
}

func (a *App) BuildUnits(builders ...UnitBuilder) error {
	for _, b := range builders {
		item, err := a.factory.Build(b.Builder)
		if err != nil {
			log.Error(err)
			return err
		}
		a.main.AddItem(item)
		a.Provide(b.Name, item)
	}
	return nil
}

func (a *App) Go(ctx context.Context) error {
	if err := a.main.Init(ctx); err != nil {
		return fmt.Errorf("init app failed: %v", err)
	}

	errc := make(chan error, 1)

	go func() {
		errc <- a.main.Start(ctx)
	}()

	const startTimeout = 10 * time.Second

	select {
	case err := <-errc:
		if err != nil {
			log.Error("Failed to start main module:", err)
			return fmt.Errorf("start app failed: %v", err)
		}
	case <-time.After(startTimeout):
		if a.main.State() != StateRunning {
			return fmt.Errorf("main module failed to start within %v", startTimeout)
		}
	}

	log.Infof("App state: %s", a.main.State())

	if err := a.wait(ctx); err != nil {
		return fmt.Errorf("app runtime error: %v", err)
	}

	return nil
}

func (a *App) wait(ctx context.Context) error {
	log.Info("Application is running. Waiting for shutdown signal...")

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	select {
	case <-ctx.Done():
		log.Warn("Context canceled, shutting down...")
	case sig := <-sigc:
		log.Infof("Received signal: %v, stopping app...", sig)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.Stop(shutdownCtx)
}

func (a *App) Stop(ctx context.Context) error {
	log.Info("Stopping main module...")
	err := a.main.Stop(ctx)
	if err != nil {
		log.Error("Failed to stop main module:", err)
		return fmt.Errorf("main module stop failed: %v", err)
	}

	log.Info("Cleaning up resources...")
	err = a.main.Cleanup(ctx)
	if err != nil {
		log.Error("Failed to clean up resources:", err)
		return fmt.Errorf("cleanup failed: %v", err)
	}

	log.Info("Application stopped successfully.")
	return nil
}

func (a *App) State() State {
	return a.main.State()
}
