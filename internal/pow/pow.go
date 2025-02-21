package pow

type Provider interface {
	Challenge(subject string, difficulty int) (string, error)
	Verify(response string) (bool, error)
}

type Solver interface {
	Solve(challenge string) (string, error)
}

type ProviderFactory interface {
	GetProvider(name string) (Provider, error)
}

type ProviderBuilder func() (Provider, error)

type Config struct {
	Difficulty int    `mapstructure:"diff" envconfig:"POW_DIFFICULTY"`
	AsyncMode  bool   `mapstructure:"async" envconfig:"POW_ASYNC"`
	RedisAddr  string `mapstructure:"redis" envconfig:"REDIS_ADDR"`
}
