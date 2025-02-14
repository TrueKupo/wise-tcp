package graceful

import (
	"context"
	"fmt"
)

type Stopper interface {
	Stop(ctx context.Context) error
}

type Service interface {
	Stopper
	fmt.Stringer
}

type Registry interface {
	Register(svc Service)
}

type Manager interface {
	Start(ctx context.Context) error
	Stopper
	Registry
}
