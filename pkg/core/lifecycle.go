package core

import "context"

type Initializer interface {
	Init(ctx context.Context) error
}

type Starter interface {
	Start(ctx context.Context) error
}

type Stopper interface {
	Stop(ctx context.Context) error
}

type Cleaner interface {
	Cleanup(ctx context.Context) error
}
