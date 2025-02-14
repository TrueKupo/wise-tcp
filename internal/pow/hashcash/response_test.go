package hashcash_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"wise-tcp/internal/pow/hashcash"
)

func TestResponseFromString(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	expiryUnix := strconv.FormatInt(expiry.Unix(), 10)
	responseStr := strings.Join([]string{
		"1",
		"20",
		expiryUnix,
		"test_subject",
		"test_nonce",
		"sha256",
		"test_solution",
	}, ":")

	resp := &hashcash.Response{}
	if err := resp.FromString(responseStr); err != nil {
		t.Fatalf("Failed to parse response string: %v", err)
	}

	if resp.Version != 1 {
		t.Errorf("Expected version 1, got %d", resp.Version)
	}
	if resp.Difficulty != 20 {
		t.Errorf("Expected difficulty 20, got %d", resp.Difficulty)
	}
	if !resp.ExpiresAt.Equal(expiry) {
		t.Errorf("Expected expiration %v, got %v", expiry, resp.ExpiresAt)
	}
	if resp.Subject != "test_subject" {
		t.Errorf("Expected subject `test_subject`, got `%s`", resp.Subject)
	}
	if resp.Nonce != "test_nonce" {
		t.Errorf("Expected nonce `test_nonce`, got `%s`", resp.Nonce)
	}
	if resp.Alg != "sha256" {
		t.Errorf("Expected algorithm `sha256`, got `%s`", resp.Alg)
	}
	if resp.Solution != "test_solution" {
		t.Errorf("Expected solution `test_solution`, got `%s`", resp.Solution)
	}
}

func TestResponseFromChallenge(t *testing.T) {
	challenge := &hashcash.Challenge{
		Payload: hashcash.Payload{
			Version:    1,
			Difficulty: 20,
			ExpiresAt:  time.Now().Add(1 * time.Minute),
			Subject:    "test_subject",
			Nonce:      "test_nonce",
			Alg:        "sha256",
		},
	}

	solution := "test_solution"
	response := &hashcash.Response{}
	response.FromChallenge(challenge, solution)

	if response.Version != challenge.Version {
		t.Errorf("Expected Version to be %d, got %d", challenge.Version, response.Version)
	}
	if response.Solution != solution {
		t.Errorf("Expected Solution to be `%s`, got `%s`", solution, response.Solution)
	}
	if response.Payload != challenge.Payload {
		t.Errorf("Expected Payload to be %+v, got %+v", challenge.Payload, response.Payload)
	}
}

func TestResponseVerify(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC()

	response := hashcash.NewResponse(
		"test_solution",
		hashcash.WithPayload(
			hashcash.Payload{
				Version:    1,
				Difficulty: 20,
				ExpiresAt:  expiry,
				Subject:    "test_subject",
				Nonce:      "test_nonce",
				Alg:        "sha256",
			},
		),
		hashcash.WithVerifier(func(hash []byte, bits, n int) (bool, error) {
			return true, nil
		}),
	)

	if response.Payload.Subject != "test_subject" {
		t.Errorf("Expected Payload.Subject to be `test_subject`, got %s", response.Payload.Subject)
	}

	if err := response.Verify(); err != nil {
		t.Errorf("Expected Verify to succeed, got error: %v", err)
	}
}

func TestResponseVerify_VerifierFails(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC()

	response := hashcash.NewResponse(
		"test_solution",
		hashcash.WithPayload(
			hashcash.Payload{
				Version:    1,
				Difficulty: 15,
				ExpiresAt:  expiry,
				Subject:    "test_subject",
				Nonce:      "test_nonce",
				Alg:        "sha256",
			},
		),
		hashcash.WithVerifier(func(hash []byte, bits, n int) (bool, error) {
			return false, nil
		}),
	)

	if err := response.Verify(); err == nil {
		t.Error("Expected an error from Verify, but got nil")
	}
}

func TestResponseVerify_UsesDefaultVerifier(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC()

	response := hashcash.NewResponse(
		"test_solution",
		hashcash.WithPayload(
			hashcash.Payload{
				Version:    1,
				Difficulty: 20,
				ExpiresAt:  expiry,
				Subject:    "test_subject",
				Nonce:      "test_nonce",
				Alg:        "sha256",
			},
		),
	)

	if err := response.Verify(); err == nil {
		t.Errorf("Expected Verify to fail because `test_solution` is invalid, but it succeeded")
	}
}
