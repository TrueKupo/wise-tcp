package providers

import (
	"fmt"
	"wise-tcp/internal/pow"
)

type Factory struct {
	builders map[string]pow.ProviderBuilder
}

func NewFactory() *Factory {
	return &Factory{
		builders: make(map[string]pow.ProviderBuilder),
	}
}

func (f *Factory) Create(name string) (pow.Provider, error) {
	if builder, ok := f.builders[name]; ok {
		return builder()
	}
	return nil, fmt.Errorf("builder for provider %s not found", name)
}

func (f *Factory) Register(name string, builder pow.ProviderBuilder) {
	f.builders[name] = builder
}
