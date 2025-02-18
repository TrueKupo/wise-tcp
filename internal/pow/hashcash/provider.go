package hashcash

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"wise-tcp/pkg/log"
)

type Provider struct {
	log        log.Logger
	cache      *Cache
	difficulty int
	expiry     time.Duration
}

type ProviderOption func(*Provider)

const defaultDifficulty = 20
const defaultExpiry = 1 * time.Minute
const defaultAlg = "sha256"

func WithDifficulty(difficulty int) ProviderOption {
	return func(s *Provider) {
		s.difficulty = difficulty
	}
}

func NewProvider(opts ...ProviderOption) *Provider {
	p := &Provider{
		difficulty: defaultDifficulty,
		expiry:     defaultExpiry,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.cache = NewCache(10 * time.Second)

	return p
}

func (p *Provider) Difficulty() int {
	return p.difficulty
}

func (p *Provider) Expiry() time.Duration {
	return p.expiry
}

func (p *Provider) Challenge(subject string, difficulty int) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); nil != err {
		return "", err
	}

	if subject == "" {
		return "", fmt.Errorf("subject must not be empty")
	}

	if difficulty == 0 {
		difficulty = p.difficulty
	} else if difficulty < 0 {
		return "", fmt.Errorf("difficulty must be positive or 0, got %d", difficulty)
	}

	c := &Challenge{
		Payload: Payload{
			Version:    1,
			Difficulty: difficulty,
			ExpiresAt:  time.Now().Add(p.expiry),
			Subject:    subject,
			Nonce:      base64.RawURLEncoding.EncodeToString(nonce),
			Alg:        defaultAlg,
		},
	}

	fingerprint, err := c.Fingerprint()
	if err != nil {
		return "", err
	}
	p.cache.Add(fingerprint, p.expiry)

	return c.String(), nil
}

func (p *Provider) Verify(response string) (bool, error) {
	r := Response{}
	if err := r.FromString(response); err != nil {
		return false, err
	}

	fingerprint, err := r.Fingerprint()
	if err != nil {
		return false, fmt.Errorf("failed to compute fingerprint: %v", err)
	}
	if err = p.cache.Remove(fingerprint); err != nil {
		return false, fmt.Errorf("replay protection failed: %v", err)
	}

	if err = r.Verify(); err != nil {
		return false, err
	}
	return true, nil
}
