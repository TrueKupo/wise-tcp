package pow

type Provider interface {
	Challenge(subject string, difficulty int) (string, error)
	Verify(response string) (bool, error)
}

type Solver interface {
	Solve(challenge string) (string, error)
}
