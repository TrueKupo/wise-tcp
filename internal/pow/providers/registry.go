package providers

import "wise-tcp/internal/pow"

type Registry struct {
	providers map[string]pow.Provider
	factory   *Factory
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]pow.Provider),
		factory:   NewFactory(),
	}
}

func (r *Registry) Register(name string, p pow.Provider) {
	r.providers[name] = p
}

func (r *Registry) RegisterBuilder(name string, builder pow.ProviderBuilder) {
	r.factory.Register(name, builder)
}

func (r *Registry) Get(name string) (pow.Provider, error) {
	p, ok := r.providers[name]
	if ok {
		return p, nil
	}
	p, err := r.factory.Create(name)
	if err != nil {
		return nil, err
	}
	r.providers[name] = p
	return p, nil
}
