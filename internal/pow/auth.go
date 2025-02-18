package pow

import (
	"context"
	"fmt"
	"io"
	"strings"
	"wise-tcp/internal/auth"
	"wise-tcp/pkg/core/build"
	"wise-tcp/pkg/log"
)

type Auth struct {
	provider Provider
}

func (a *Auth) AuthorizeRequest(ctx context.Context, request auth.Request, rw io.ReadWriter) error {
	subject := request.ClientAddr

	// Generate the challenge
	challenge, err := a.provider.Challenge(subject, 0)
	if err != nil {
		return fmt.Errorf("failed to generate challenge: %w", err)
	}

	// Write the challenge
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
		return fmt.Errorf("context canceled during challenge write: %w", ctx.Err())
	case err := <-writeDone:
		if err != nil {
			return err
		}
	}

	// Read response
	responseDone := make(chan error, 1)
	var response []byte
	go func() {
		response = make([]byte, 128)
		_, err := rw.Read(response)
		responseDone <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during response read: %w", ctx.Err())
	case err := <-responseDone:
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	// Verify solution
	solution := strings.TrimSpace(strings.TrimPrefix(string(response), "X-Response:"))
	verifyDone := make(chan error, 1)
	var valid bool
	go func() {
		var err error
		valid, err = a.provider.Verify(solution)
		verifyDone <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during verification: %w", ctx.Err())
	case err := <-verifyDone:
		if err != nil {
			return err
		}
	}

	if !valid {
		// todo: configure by server policy
		_, err := rw.Write([]byte("X-Err: invalid solution\n"))
		if err != nil {
			log.Error(err)
		}
		return auth.ErrUnauthorized
	}

	return nil
}

func AuthBuilder() build.Builder {
	return func(i *build.Injector) (any, error) {
		provider, ok := i.Get("auth.provider").(Provider)
		if !ok {
			return nil, fmt.Errorf("auth.provider not found")
		}
		return NewAuth(provider), nil
	}
}

func NewAuth(provider Provider) *Auth {
	return &Auth{
		provider: provider,
	}
}
