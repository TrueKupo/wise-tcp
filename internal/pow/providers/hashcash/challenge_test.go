package hashcash_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"wise-tcp/internal/pow/providers/hashcash"
)

func TestChallengeFromString(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	expiryUnix := strconv.FormatInt(expiry.Unix(), 10)
	challengeStr := strings.Join([]string{
		"1",
		"20",
		expiryUnix,
		"test_subject",
		"test_nonce",
		"sha256",
	}, ":")

	ch := &hashcash.Challenge{}
	if err := ch.FromString(challengeStr); err != nil {
		t.Fatalf("Failed to parse challenge string: %v", err)
	}

	if ch.Version != 1 {
		t.Errorf("Expected version 1, got %d", ch.Version)
	}
	if ch.Difficulty != 20 {
		t.Errorf("Expected difficulty 20, got %d", ch.Difficulty)
	}
	if !ch.ExpiresAt.Equal(expiry) {
		t.Errorf("Expected expiration %v, got %v", expiry, ch.ExpiresAt)
	}
	if ch.Subject != "test_subject" {
		t.Errorf("Expected subject `test_subject`, got `%s`", ch.Subject)
	}
	if ch.Nonce != "test_nonce" {
		t.Errorf("Expected nonce `test_nonce`, got `%s`", ch.Nonce)
	}
	if ch.Alg != "sha256" {
		t.Errorf("Expected algorithm `sha256`, got `%s`", ch.Alg)
	}
}

func TestChallengeRoundTrip(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	payload := &hashcash.Payload{
		Version:    1,
		Difficulty: 20,
		ExpiresAt:  expiry,
		Subject:    "test_subject",
		Nonce:      "test_nonce",
		Alg:        "sha256",
	}
	challenge := &hashcash.Challenge{Payload: *payload}

	serialized := challenge.Payload.String()
	deserialized := &hashcash.Challenge{}
	if err := deserialized.FromString(serialized); err != nil {
		t.Fatalf("Failed to deserialize challenge: %v", err)
	}

	if challenge.Payload != deserialized.Payload {
		t.Errorf("Mismatch in Challenge during roundtrip: %+v vs %+v", challenge.Payload, deserialized.Payload)
	}
}
