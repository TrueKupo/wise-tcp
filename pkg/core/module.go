package core

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"wise-tcp/pkg/log"
)

type Module struct {
	name  string
	cfg   interface{}
	mods  map[string]*Module
	units []*Unit
	state *stateLock
}

type ModFactory[C ModConfig] func(cfg C) (*Module, error)

type ModConfig interface {
	Name() string
}

func NewModule(name string) *Module {
	return &Module{
		name:  name,
		mods:  make(map[string]*Module),
		state: &stateLock{state: StateNone, verbose: true},
	}
}

func (m *Module) AddItem(item interface{}) *Module {
	unit := NewUnit(item)
	m.units = append(m.units, unit)
	return m
}

func (m *Module) AddModule(mod *Module) *Module {
	m.mods[mod.Name()] = mod
	return m
}

func (m *Module) GetModule(name string) *Module {
	return m.mods[name]
}

func (m *Module) Init(ctx context.Context) error {
	if m.state.Get() != StateNone {
		return errors.New("module is already initialized or in invalid state")
	}
	m.state.Set(StateInit)

	for _, mod := range m.mods {
		if err := mod.Init(ctx); err != nil {
			log.Error("Failed to initialize module:", mod.Name(), err)
			m.state.Set(StateError)
			return err
		}
	}

	for _, unit := range m.units {
		if err := unit.Init(ctx); err != nil {
			log.Error("Failed to initialize unit:", err)
			m.state.Set(StateError)
			return err
		}
	}

	m.state.Set(StateReady)
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	if m.state.Get() != StateReady {
		return errors.New("module must be in 'Ready' state to start")
	}
	m.state.Set(StateStarting)

	var wg sync.WaitGroup
	errc := make(chan error, len(m.mods)+len(m.units))

	for _, mod := range m.mods {
		wg.Add(1)
		go func(mod *Module) {
			defer wg.Done()
			if err := mod.Start(ctx); err != nil {
				errc <- fmt.Errorf("module [%s] failed to start: %w", mod.Name(), err)
			}
		}(mod)
	}

	for _, unit := range m.units {
		wg.Add(1)
		go func(unit *Unit) {
			defer wg.Done()
			if err := unit.Start(ctx); err != nil {
				errc <- fmt.Errorf("unit failed to start: %w", err)
			}
		}(unit)
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	var firstErr error
	for err := range errc {
		if firstErr == nil {
			firstErr = err
		}
		log.Error("Error during start:", err)
	}

	if firstErr != nil {
		m.state.Set(StateError)
		return firstErr
	}

	m.state.Set(StateRunning)
	return nil
}

func (m *Module) Stop(ctx context.Context) error {
	if m.state.Get() != StateRunning {
		return errors.New("module must be in 'Running' state to stop")
	}
	m.state.Set(StateStopping)

	var wg sync.WaitGroup
	errc := make(chan error, len(m.units)+len(m.mods))
	done := make(chan struct{})

	for _, mod := range m.mods {
		wg.Add(1)
		go func(mod *Module) {
			defer wg.Done()
			if err := mod.Stop(ctx); err != nil {
				errc <- fmt.Errorf("module [%s] failed to stop: %w", mod.Name(), err)
			}
		}(mod)
	}

	for _, unit := range m.units {
		wg.Add(1)
		go func(unit *Unit) {
			defer wg.Done()
			if err := unit.Stop(ctx); err != nil {
				errc <- fmt.Errorf("unit failed to stop: %w", err)
			}
		}(unit)
	}

	go func() {
		wg.Wait()
		close(done)
		close(errc)
	}()

	var firstErr error
	for {
		select {
		case <-ctx.Done():
			log.Warn("Stop operation exceeded context deadline or canceled")
			return fmt.Errorf("stop operation timed out: %w", ctx.Err())

		case <-done:
			if firstErr != nil {
				m.state.Set(StateError)
				return firstErr
			}
			m.state.Set(StateStopped)
			return nil

		case err := <-errc:
			if firstErr == nil {
				firstErr = err
			}
			log.Error("Error during module stop:", err)
		}
	}
}

func (m *Module) Cleanup(ctx context.Context) error {
	if m.state.Get() != StateStopped {
		return errors.New("module must be in 'Stopped' state to clean up")
	}
	m.state.Set(StateCleanup)

	for _, unit := range m.units {
		if err := unit.Cleanup(ctx); err != nil {
			log.Error("Unit cleanup failed:", err)
			m.state.Set(StateError)
			return err
		}
	}

	for _, mod := range m.mods {
		if err := mod.Cleanup(ctx); err != nil {
			log.Error("Module cleanup failed:", mod.Name(), err)
			m.state.Set(StateError)
			return err
		}
	}

	m.state.Set(StateFinished)
	return nil
}

func (m *Module) String() string {
	return m.name
}

func (m *Module) Name() string {
	return m.name
}

func (m *Module) State() State {
	return m.state.Get()
}
