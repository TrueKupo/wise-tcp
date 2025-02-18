package core

import (
	"sync"
	"wise-tcp/pkg/log"
)

type State uint8

const (
	StateNone State = iota
	StateCreated
	StateInit
	StateReady
	StateStarting
	StateRunning
	StateStopping
	StateStopped
	StateCleanup
	StateFinished
	StateError
)

type stateLock struct {
	sync.Mutex
	state   State
	verbose bool
}

func (l *stateLock) Set(state State) {
	l.Lock()
	defer l.Unlock()
	if l.verbose {
		log.Debugf("state transition: %s -> %s", l.state, state)
	}
	l.state = state
}

func (l *stateLock) Get() State {
	return l.state
}

func (s State) String() string {
	switch s {
	case StateNone:
		return "None"
	case StateCreated:
		return "Created"
	case StateInit:
		return "Init"
	case StateReady:
		return "Ready"
	case StateStarting:
		return "Starting"
	case StateRunning:
		return "Running"
	case StateStopping:
		return "Stopping"
	case StateStopped:
		return "Stopped"
	case StateCleanup:
		return "Cleanup"
	case StateFinished:
		return "Finished"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}
