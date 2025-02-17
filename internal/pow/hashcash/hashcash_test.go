package hashcash_test

import (
	"testing"
	"time"

	"wise-tcp/internal/pow/hashcash"
	fc "wise-tcp/pkg/factory"
	"wise-tcp/pkg/log"
)

func TestProvider_Challenge(t *testing.T) {
	factory := fc.New(fc.WithLogger(log.Default()))
	provider := hashcash.NewProvider(factory)

	subject := "test_subject"
	difficulty := 20

	challenge, err := provider.Challenge(subject, difficulty)
	if err != nil {
		t.Fatalf("Failed to create challenge: %v", err)
	}

	ch := &hashcash.Challenge{}
	if err := ch.FromString(challenge); err != nil {
		t.Errorf("Invalid challenge format: %v", err)
	}

	if ch.Subject != subject {
		t.Errorf("Expected subject %s, got %s", subject, ch.Subject)
	}

	if ch.Difficulty != difficulty {
		t.Errorf("Expected difficulty %d, got %d", difficulty, ch.Difficulty)
	}

	if ch.ExpiresAt.Before(time.Now()) {
		t.Errorf("Challenge expiration time is invalid: %v", ch.ExpiresAt)
	}
}

func TestProvider_Verify(t *testing.T) {
	factory := fc.New(fc.WithLogger(log.Default()))
	provider := hashcash.NewProvider(factory)

	subject := "test_subject"
	difficulty := 20

	challenge, err := provider.Challenge(subject, difficulty)
	if err != nil {
		t.Fatalf("Failed to create challenge: %v", err)
	}

	solver := hashcash.NewSolver()
	response, err := solver.Solve(challenge)
	if err != nil {
		t.Fatalf("Failed to solve challenge: %v", err)
	}

	result, err := provider.Verify(response)
	if err != nil {
		t.Fatalf("Failed to verify response: %v", err)
	}

	if !result {
		t.Error("Expected valid solution, but verification failed")
	}

	result, err = provider.Verify(response)
	if err == nil {
		t.Errorf("Replay protection failed: Expected error but got success")
	}
	if result {
		t.Errorf("Replay protection failed: Solution should not pass twice")
	}
}

func TestProvider_Verify_FakeSolution(t *testing.T) {
	factory := fc.New(fc.WithLogger(log.Default()))
	provider := hashcash.NewProvider(factory)

	subject := "test_subject"
	difficulty := 20

	validChallenge, err := provider.Challenge(subject, difficulty)
	if err != nil {
		t.Fatalf("Failed to create challenge: %v", err)
	}

	fakeChallenge := &hashcash.Challenge{
		Payload: hashcash.Payload{
			Version:    1,
			Difficulty: difficulty,
			ExpiresAt:  time.Now().Add(1 * time.Minute),
			Subject:    subject,
			Nonce:      "fake_nonce",
			Alg:        "sha256",
		},
	}

	solver := hashcash.NewSolver()

	fakeSolution, err := solver.GetSolution(fakeChallenge.String())
	if err != nil {
		t.Fatalf("Failed to solve fake challenge: %v", err)
	}

	validSolution, err := solver.GetSolution(validChallenge)
	if err != nil {
		t.Fatalf("Failed to solve valid challenge: %v", err)
	}

	r, err := hashcash.ResponseFromChallenge(validChallenge, fakeSolution)
	if err != nil {
		t.Fatalf("Failed to construct response: %v", err)
	}
	err = r.Verify()
	if err == nil {
		t.Errorf("Expected error for invalid solution, but got no error")
	}

	r, err = hashcash.ResponseFromChallenge(validChallenge, validSolution)
	if err != nil {
		t.Fatalf("Failed to construct response: %v", err)
	}
	err = r.Verify()
	if err != nil {
		t.Errorf("Expected success for valid solution, but got error: %v", err)
	}
}

func TestSolver_Solve(t *testing.T) {
	subject := "test_subject"
	difficulty := 20
	expiry := time.Now().Add(1 * time.Minute)

	ch := &hashcash.Challenge{
		Payload: hashcash.Payload{
			Version:    1,
			Difficulty: difficulty,
			ExpiresAt:  expiry,
			Subject:    subject,
			Nonce:      "nonce",
			Alg:        "sha256",
		},
	}

	chStr := ch.String()
	solver := hashcash.NewSolver()

	response, err := solver.Solve(chStr)
	if err != nil {
		t.Fatalf("Failed to solve challenge: %v", err)
	}

	r := &hashcash.Response{}
	if err := r.FromString(response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if err := r.Verify(); err != nil {
		t.Errorf("Invalid solution: %v", err)
	}

	if r.Subject != ch.Subject {
		t.Errorf("Expected subject %s, got %s", ch.Subject, r.Subject)
	}

	if r.Difficulty != ch.Difficulty {
		t.Errorf("Expected difficulty %d, got %d", ch.Difficulty, r.Difficulty)
	}
}

func TestCache_ReplayProtection(t *testing.T) {
	factory := fc.New(fc.WithLogger(log.Default()))
	provider := hashcash.NewProvider(factory)

	subject := "test_subject"
	difficulty := 10

	challenge, err := provider.Challenge(subject, difficulty)
	if err != nil {
		t.Fatalf("Failed to create challenge: %v", err)
	}

	solver := hashcash.NewSolver()
	response, err := solver.Solve(challenge)
	if err != nil {
		t.Fatalf("Failed to solve challenge: %v", err)
	}

	result, err := provider.Verify(response)
	if err != nil {
		t.Fatalf("Failed to verify response: %v", err)
	}
	if !result {
		t.Error("Expected valid solution, but verification failed")
	}

	result, err = provider.Verify(response)
	if err == nil {
		t.Error("Expected replay protection error, got no error")
	}
	if result {
		t.Error("Expected replay protection to invalidate the solution")
	}
}
