package pow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"wise-tcp/internal/auth"
	"wise-tcp/internal/pow/providers/hashcash"
	"wise-tcp/pkg/core"
	"wise-tcp/pkg/core/build"
	"wise-tcp/pkg/log"
)

type Auth struct {
	provider Provider
	async    bool
}

func AuthBuilder(cfg Config) build.Builder {
	return func(_ *build.Injector) (any, error) {
		opts := []hashcash.ProviderOption{
			hashcash.WithDifficulty(cfg.Difficulty),
		}
		if cfg.AsyncMode {
			opts = append(opts, hashcash.WithCache(hashcash.NewRedisCache(cfg.RedisAddr)))
		}
		provider := hashcash.NewProvider(opts...)
		return NewAuth(provider, cfg.AsyncMode), nil
	}
}

func NewAuth(provider Provider, async bool) *Auth {
	return &Auth{
		provider: provider,
		async:    async,
	}
}

func (a *Auth) Start(ctx context.Context) error {
	if starter, ok := a.provider.(core.Starter); ok {
		return starter.Start(ctx)
	}
	return nil
}

func (a *Auth) Stop(ctx context.Context) error {
	if starter, ok := a.provider.(core.Stopper); ok {
		return starter.Stop(ctx)
	}
	return nil
}

func (a *Auth) AuthorizeRequest(ctx context.Context, request auth.Request, rw io.ReadWriter) error {
	if a.async {
		return a.handleAsyncMode(ctx, rw)
	}

	return a.handleSyncMode(ctx, request.ClientAddr, rw)
}

func (a *Auth) handleSyncMode(ctx context.Context, subject string, rw io.ReadWriter) error {
	challenge, err := a.provider.Challenge(subject, 0)
	if err != nil {
		return fmt.Errorf("failed to generate challenge: %w", err)
	}

	if err := a.sendChallenge(ctx, rw, challenge); err != nil {
		return err
	}

	response, err := a.readResponse(ctx, rw)
	if err != nil {
		return err
	}

	solution, ok := a.parseResponse(response)
	if !ok {
		return auth.ErrProtoMismatch
	}

	return a.verifySolution(ctx, solution, rw)
}

func (*Auth) parseResponse(response []byte) (string, bool) {
	if !strings.HasPrefix(string(response), "X-Response:") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(string(response), "X-Response:")), true
}

func (a *Auth) handleAsyncMode(ctx context.Context, rw io.ReadWriter) error {
	readCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	response, err := a.readResponse(readCtx, rw)
	if err != nil {
		return err
	}

	solution, ok := a.parseResponse(response)
	if !ok {
		return auth.ErrProtoMismatch
	}

	return a.verifySolution(ctx, solution, rw)
}

func (a *Auth) sendChallenge(ctx context.Context, rw io.Writer, challenge string) error {
	writeDone := make(chan error, 1)
	go func() {
		writeDone <- func() error {
			_, err := rw.Write([]byte(fmt.Sprintf("X-Challenge: %s\n", challenge)))
			if err != nil {
				return fmt.Errorf("failed to write challenge: %w", err)
			}
			return nil
		}()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-writeDone:
		return err
	}
}

func (a *Auth) readResponse(ctx context.Context, rw io.Reader) ([]byte, error) {
	responseDone := make(chan error, 1)
	response := make([]byte, 128)

	var n int
	go func() {
		var err error
		n, err = rw.Read(response)
		responseDone <- err
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-responseDone:
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
	}

	return bytes.Trim(response[:n], " \n\x00"), nil
}

func (a *Auth) verifySolution(ctx context.Context, solution string, rw io.Writer) error {
	verifyDone := make(chan error, 1)
	var valid bool

	go func() {
		var err error
		valid, err = a.provider.Verify(solution)
		verifyDone <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-verifyDone:
		if err != nil {
			return fmt.Errorf("verification error: %w", err)
		}
	}

	if !valid {
		_, err := rw.Write([]byte("X-Err: invalid solution\n"))
		if err != nil {
			log.Error(err)
		}
		return auth.ErrUnauthorized
	}

	return nil
}
