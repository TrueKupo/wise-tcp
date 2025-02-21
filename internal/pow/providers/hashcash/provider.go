package hashcash

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"wise-tcp/pkg/core"
)

type Provider struct {
	cache      ChallengeCache
	difficulty int
	expiry     time.Duration
}

type ProviderOption func(*Provider)

const defaultDifficulty = 20
const defaultExpiry = 1 * time.Minute
const defaultAlg = "sha256"

type Config struct {
	Difficulty int
}

type ChallengeCache interface {
	Add(fingerprint string, challenge string, expiration time.Duration) error
	Remove(fingerprint string) error
	core.Starter
	core.Stopper
}

func WithDifficulty(difficulty int) ProviderOption {
	return func(s *Provider) {
		s.difficulty = difficulty
	}
}

func WithCache(cache ChallengeCache) ProviderOption {
	return func(provider *Provider) {
		provider.cache = cache
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

	if p.cache == nil {
		p.cache = NewMemoryCache(10 * time.Second)
	}

	return p
}

func (p *Provider) Difficulty() int {
	return p.difficulty
}

func (p *Provider) Expiry() time.Duration {
	return p.expiry
}

func (p *Provider) Start(ctx context.Context) error {
	return p.cache.Start(ctx)
}

func (p *Provider) Stop(ctx context.Context) error {
	return p.cache.Stop(ctx)
}

func (p *Provider) Challenge(subject string, difficulty int) (string, error) {
	subject = base64.RawURLEncoding.EncodeToString([]byte(subject))
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

	err = p.cache.Add(fingerprint, c.String(), p.expiry)
	if err != nil {
		return "", err
	}

	return c.String(), nil
}

func (p *Provider) RawChallenge(subject string, difficulty int) (*Challenge, error) {
	subject = base64.RawURLEncoding.EncodeToString([]byte(subject))
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); nil != err {
		return nil, err
	}

	if subject == "" {
		return nil, fmt.Errorf("subject must not be empty")
	}

	if difficulty == 0 {
		difficulty = p.difficulty
	} else if difficulty < 0 {
		return nil, fmt.Errorf("difficulty must be positive or 0, got %d", difficulty)
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

	return c, nil
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
		if errors.Is(err, ErrInvalidSolution) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
