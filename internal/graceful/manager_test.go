package graceful

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"wise-tcp/pkg/log"
)

type MockStopper struct {
	Name      string
	StopError error
	Stopped   bool
	StopDelay time.Duration
}

func (m *MockStopper) String() string {
	return m.Name
}

func (m *MockStopper) Stop(ctx context.Context) error {
	m.Stopped = true
	if m.StopDelay > 0 {
		select {
		case <-time.After(m.StopDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return m.StopError
}

func TestManager_Stop(t *testing.T) {
	logger := log.Default()

	svc1 := &MockStopper{Name: "Service1"}
	svc2 := &MockStopper{Name: "Service2", StopError: fmt.Errorf("failed to stop")}

	mgr := NewManager(WithLogger(logger), WithTimeout(5*time.Second))
	mgr.Register(svc1)
	mgr.Register(svc2)

	err := mgr.Stop(context.Background())

	assert.Error(t, err)
	assert.True(t, svc1.Stopped)
	assert.True(t, svc2.Stopped)
	assert.Contains(t, err.Error(), "failed to stop")
}
