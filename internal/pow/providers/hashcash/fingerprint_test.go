package hashcash

import (
	"strings"
	"testing"
	"time"
)

func TestGetFingerprint(t *testing.T) {
	tests := []struct {
		name        string
		alg         string
		subject     string
		nonce       string
		at          time.Time
		difficulty  int
		expected    string
		expectError bool
	}{
		{
			name:        "Valid inputs",
			alg:         "SHA256",
			subject:     "testSubject",
			nonce:       "12345",
			at:          time.Now().Add(10 * time.Second),
			difficulty:  5,
			expected:    "SHA256:testSubject:12345:",
			expectError: false,
		},
		{
			name:        "Empty algorithm",
			alg:         "",
			subject:     "testSubject",
			nonce:       "12345",
			at:          time.Now().Add(10 * time.Second),
			difficulty:  5,
			expectError: true,
		},
		{
			name:        "Empty subject",
			alg:         "SHA256",
			subject:     "",
			nonce:       "12345",
			at:          time.Now().Add(10 * time.Second),
			difficulty:  5,
			expectError: true,
		},
		{
			name:        "Empty nonce",
			alg:         "SHA256",
			subject:     "testSubject",
			nonce:       "",
			at:          time.Now().Add(10 * time.Second),
			difficulty:  5,
			expectError: true,
		},
		{
			name:        "Negative difficulty",
			alg:         "SHA256",
			subject:     "testSubject",
			nonce:       "12345",
			at:          time.Now().Add(10 * time.Second),
			difficulty:  -1,
			expectError: true,
		},
		{
			name:        "Zero difficulty",
			alg:         "SHA256",
			subject:     "testSubject",
			nonce:       "12345",
			at:          time.Now().Add(10 * time.Second),
			difficulty:  0,
			expectError: true,
		},
		{
			name:        "Timestamp in the past",
			alg:         "SHA256",
			subject:     "testSubject",
			nonce:       "12345",
			at:          time.Now().Add(-10 * time.Second),
			difficulty:  5,
			expectError: true,
		},
		{
			name:        "Whitespace trimmed inputs",
			alg:         "  SHA256 ",
			subject:     " testSubject ",
			nonce:       " 12345 ",
			at:          time.Now().Add(10 * time.Second),
			difficulty:  5,
			expected:    "SHA256:testSubject:12345:",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getFingerprint(tt.alg, tt.subject, tt.nonce, tt.at, tt.difficulty)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
				return
			}

			if err == nil && !strings.HasPrefix(result, tt.expected) {
				t.Errorf("Expected result prefix: %v, got: %v", tt.expected, result)
			}
		})
	}
}
