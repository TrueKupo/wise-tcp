package core

import (
	"context"
	"errors"
	"fmt"
	"wise-tcp/pkg/log"
)

type Unit struct {
	state stateLock
	item  interface{}
}

func NewUnit(item interface{}) *Unit {
	return &Unit{
		state: stateLock{state: StateNone},
		item:  item,
	}
}

func (u *Unit) Init(ctx context.Context) error {
	if u.state.Get() != StateNone {
		return errors.New("unit already initialized or in invalid state")
	}

	u.state.Set(StateInit)

	if initializer, ok := u.item.(Initializer); ok {
		if err := initializer.Init(ctx); err != nil {
			u.state.Set(StateError)
			return fmt.Errorf("failed to initialize unit: %w", err)
		}
	}

	u.state.Set(StateReady)
	return nil
}

func (u *Unit) Start(ctx context.Context) error {
	if u.state.Get() != StateReady {
		return errors.New("unit is not ready to start")
	}

	if ctx.Err() != nil {
		return errors.New("context canceled before unit start")
	}

	u.state.Set(StateStarting)

	if starter, ok := u.item.(Starter); ok {
		if err := starter.Start(ctx); err != nil {
			u.state.Set(StateError)
			return fmt.Errorf("failed to start unit: %w", err)
		}
	}

	u.state.Set(StateRunning)
	return nil
}

func (u *Unit) Stop(ctx context.Context) error {
	if u.state.Get() != StateRunning {
		return errors.New("unit is not in 'Running' state")
	}

	u.state.Set(StateStopping)

	if stopper, ok := u.item.(Stopper); ok {
		errc := make(chan error, 1)
		go func() {
			errc <- stopper.Stop(ctx)
		}()

		select {
		case <-ctx.Done():
			u.state.Set(StateError)
			return fmt.Errorf("timeout or context canceled during unit stop")
		case err := <-errc:
			if err != nil {
				u.state.Set(StateError)
				return fmt.Errorf("failed to stop unit: %w", err)
			}
		}
	}

	log.Info("Unit stopped")

	u.state.Set(StateStopped)
	return nil
}

func (u *Unit) Cleanup(ctx context.Context) error {
	if u.state.Get() != StateStopped {
		return errors.New("unit must be in 'Stopped' state before cleanup")
	}

	u.state.Set(StateCleanup)

	if cleaner, ok := u.item.(Cleaner); ok {
		if err := cleaner.Cleanup(ctx); err != nil {
			u.state.Set(StateError)
			return fmt.Errorf("failed to cleanup unit: %w", err)
		}
	}

	u.state.Set(StateFinished)
	return nil
}

func (u *Unit) State() State {
	return u.state.Get()
}
