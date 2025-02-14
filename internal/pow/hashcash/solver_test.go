package hashcash_test

import (
	"testing"
	"time"
	"wise-tcp/internal/pow/hashcash"
)

func TestSolver_Solve_ValidChallenge(t *testing.T) {
	solver := hashcash.NewSolver()

	expiry := time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	challenge := &hashcash.Challenge{
		Payload: hashcash.Payload{
			Version:    1,
			Difficulty: 10,
			ExpiresAt:  expiry,
			Subject:    "test_subject",
			Nonce:      "test_nonce",
			Alg:        "SHA256",
		},
	}

	solution, err := solver.Solve(challenge.String())
	if err != nil {
		t.Fatalf("failed to solve challenge: %v", err)
	}

	response := &hashcash.Response{}
	err = response.FromString(solution)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if err := response.Verify(); err != nil {
		t.Fatalf("solver produced an invalid solution")
	}
}

func TestSolver_Solve_InvalidChallenge(t *testing.T) {
	solver := hashcash.NewSolver()

	invalidChallengeStr := "subject"

	_, err := solver.Solve(invalidChallengeStr)
	if err == nil {
		t.Fatal("expected error when solving invalid challenge, but got nil")
	}
}

func TestSolver_Solve_HighDifficulty(t *testing.T) {
	solver := hashcash.NewSolver()

	expiry := time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	challenge := &hashcash.Challenge{
		Payload: hashcash.Payload{
			Version:    1,
			Difficulty: 10,
			ExpiresAt:  expiry,
			Subject:    "test_subject",
			Nonce:      "test_nonce",
			Alg:        "SHA256",
		},
	}

	solution, err := solver.Solve(challenge.String())
	if err != nil {
		t.Fatalf("failed to solve high-difficulty challenge: %v", err)
	}

	response := &hashcash.Response{}
	err = response.FromString(solution)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if err = response.Verify(); err != nil {
		t.Fatalf("solver produced an invalid solution for high-difficulty challenge")
	}
}

func TestSolver_Solve_NoSolution(t *testing.T) {
	solver := hashcash.NewSolver()

	impossibleChallengeStr := "subject:0"

	_, err := solver.Solve(impossibleChallengeStr)
	if err == nil {
		t.Fatal("expected error for impossible challenge, but got nil")
	}
}
