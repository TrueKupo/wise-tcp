package hashcash_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"wise-tcp/internal/pow/hashcash"
)

func TestPayloadStringSerialization(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC()
	payload := &hashcash.Payload{
		Version:    1,
		Difficulty: 20,
		ExpiresAt:  expiry,
		Subject:    "test_subject",
		Nonce:      "test_nonce",
		Alg:        "sha256",
	}

	expected := payload.String()
	parts := []string{
		"1",
		"20",
		strconv.FormatInt(expiry.Unix(), 10),
		"test_subject",
		"test_nonce",
		"sha256",
	}
	actual := strings.Join(parts, ":")
	if expected != actual {
		t.Errorf("Expected payload string `%s`, got `%s`", actual, expected)
	}
}

func TestPayloadFromString(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	expiryUnix := strconv.FormatInt(expiry.Unix(), 10)
	payloadStr := strings.Join([]string{"1", "20", expiryUnix, "test_subject", "test_nonce", "sha256"}, ":")

	payload := &hashcash.Payload{}
	if err := payload.FromString(strings.Split(payloadStr, ":")); err != nil {
		t.Fatalf("Failed to parse payload string: %v", err)
	}

	if payload.Version != 1 {
		t.Errorf("Expected version 1, got %d", payload.Version)
	}
	if payload.Difficulty != 20 {
		t.Errorf("Expected difficulty 20, got %d", payload.Difficulty)
	}
	if !payload.ExpiresAt.Equal(expiry) {
		t.Errorf("Expected expiration %v, got %v", expiry, payload.ExpiresAt)
	}
	if payload.Subject != "test_subject" {
		t.Errorf("Expected subject `test_subject`, got `%s`", payload.Subject)
	}
	if payload.Nonce != "test_nonce" {
		t.Errorf("Expected nonce `test_nonce`, got `%s`", payload.Nonce)
	}
	if payload.Alg != "sha256" {
		t.Errorf("Expected algorithm `sha256`, got `%s`", payload.Alg)
	}
}

func TestPayloadFromStringInvalidCases(t *testing.T) {
	tests := []struct {
		name      string
		parts     []string
		expectErr bool
	}{
		{"TooFewParts", []string{"1", "20"}, true},
		{"InvalidVersion", []string{"x", "20", "0", "subj", "nonce", "sha256"}, true},
		{"InvalidDifficulty", []string{"1", "-1", "0", "subj", "nonce", "sha256"}, true},
		{"ExpiredTime", []string{"1", "20", "0", "subj", "nonce", "sha256"}, true},
		{"EmptySubject", []string{"1", "20", "9999999999", "", "nonce", "sha256"}, true},
		{"EmptyNonce", []string{"1", "20", "9999999999", "subj", "", "sha256"}, true},
		{"EmptyAlg", []string{"1", "20", "9999999999", "subj", "nonce", ""}, true},
		{"ValidPayload", []string{"1", "20", "9999999999", "subj", "nonce", "sha256"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &hashcash.Payload{}
			err := payload.FromString(tt.parts)
			if hasErr := err != nil; hasErr != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestPayloadFingerprint(t *testing.T) {
	payload1 := &hashcash.Payload{
		Version:    1,
		Difficulty: 20,
		ExpiresAt:  time.Now().Add(1 * time.Minute).UTC(),
		Subject:    "test_1",
		Nonce:      "nonce_1",
		Alg:        "sha256",
	}

	payload2 := &hashcash.Payload{
		Version:    1,
		Difficulty: 20,
		ExpiresAt:  time.Now().Add(1 * time.Minute).UTC(),
		Subject:    "test_2",
		Nonce:      "nonce_2",
		Alg:        "sha256",
	}

	f1, err1 := payload1.Fingerprint()
	f2, err2 := payload2.Fingerprint()

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to generate fingerprints: %v, %v", err1, err2)
	}

	if f1 == f2 {
		t.Errorf("Expected different fingerprints for different payloads, got same: %s", f1)
	}
}

func TestPayloadExpirationValidation(t *testing.T) {
	validExpiry := strings.Join([]string{"1", "20", strconv.FormatInt(time.Now().Add(1*time.Minute).Unix(), 10), "test_subject", "test_nonce", "sha256"}, ":")
	expiredExpiry := strings.Join([]string{"1", "20", strconv.FormatInt(time.Now().Add(-1*time.Minute).Unix(), 10), "test_subject", "test_nonce", "sha256"}, ":")

	payload := &hashcash.Payload{}

	if err := payload.FromString(strings.Split(validExpiry, ":")); err != nil {
		t.Errorf("Unexpected error for valid expiration: %v", err)
	}

	if err := payload.FromString(strings.Split(expiredExpiry, ":")); err == nil {
		t.Error("Expected error for expired payload, got none")
	}
}

func TestPayloadStringRoundTrip(t *testing.T) {
	expiry := time.Now().Add(1 * time.Minute).UTC().Truncate(time.Second)
	payload1 := &hashcash.Payload{
		Version:    1,
		Difficulty: 10,
		ExpiresAt:  expiry,
		Subject:    "test_subject",
		Nonce:      "test_nonce",
		Alg:        "sha256",
	}

	asString := payload1.String()
	payload2 := &hashcash.Payload{}
	if err := payload2.FromString(strings.Split(asString, ":")); err != nil {
		t.Fatalf("Failed to parse serialized payload: %v", err)
	}

	if *payload1 != *payload2 {
		t.Errorf("Expected payloads to match, got: %+v, %+v", payload1, payload2)
	}
}
