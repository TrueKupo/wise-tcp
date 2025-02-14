package pow

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
)

type GuardConfig struct {
	PowDifficulty int `yaml:"powDifficulty" env:"POW_DIFFICULTY"`
}

type NetGuard struct {
	provider Provider
}

func NewGuard(provider Provider) *NetGuard {
	return &NetGuard{
		provider: provider,
	}
}

func (g *NetGuard) Verify(_ context.Context, conn net.Conn) error {
	subject := base64.StdEncoding.EncodeToString([]byte(conn.RemoteAddr().String()))
	challenge, err := g.provider.Challenge(subject, 0)
	if err != nil {
		return fmt.Errorf("failed to generate challenge: %w", err)
	}

	_, err = conn.Write([]byte(fmt.Sprintf("X-Challenge: %s\n", challenge)))
	if err != nil {
		return fmt.Errorf("failed to write challenge: %w", err)
	}

	response := make([]byte, 128)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	solution := strings.TrimSpace(strings.TrimPrefix(string(response[:n]), "X-Response:"))

	valid, err := g.provider.Verify(solution)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid solution")
	}

	return nil
}
