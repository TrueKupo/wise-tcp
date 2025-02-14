package hashcash

import (
	"testing"
)

func TestVerifyBits(t *testing.T) {
	tests := []struct {
		name        string
		hash        []byte
		bits        int
		n           int
		expected    bool
		expectError bool
	}{
		{
			name:        "Valid bits, full zero bytes",
			hash:        []byte{0x00, 0x00, 0x00, 0x00},
			bits:        24,
			n:           3,
			expected:    true,
			expectError: false,
		},
		{
			name:        "Valid bits, mixed bytes with remaining bits",
			hash:        []byte{0x00, 0x00, 0x7F, 0x00},
			bits:        17,
			n:           3,
			expected:    true,
			expectError: false,
		},
		{
			name:        "Bits not satisfied",
			hash:        []byte{0x00, 0xFF, 0x00, 0x00},
			bits:        16,
			n:           3,
			expected:    false,
			expectError: false,
		},
		{
			name:        "Remaining bits not satisfied",
			hash:        []byte{0x00, 0x00, 0x80, 0x00},
			bits:        17,
			n:           3,
			expected:    false,
			expectError: false,
		},
		{
			name:        "Invalid bits - negative",
			hash:        []byte{0x00, 0x00, 0x00},
			bits:        -1,
			n:           3,
			expectError: true,
		},
		{
			name:        "Invalid bits - exceeds hash size",
			hash:        []byte{0x00, 0x00, 0x00},
			bits:        64,
			n:           3,
			expectError: true,
		},
		{
			name:        "n exceeds hash length",
			hash:        []byte{0x00, 0x00},
			bits:        16,
			n:           10,
			expected:    true,
			expectError: false,
		},
		{
			name:        "Empty hash",
			hash:        []byte{},
			bits:        8,
			n:           1,
			expectError: true,
		},
		{
			name:        "Zero bits",
			hash:        []byte{0x00, 0x00},
			bits:        0,
			n:           2,
			expected:    true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := verifyBits(tt.hash, tt.bits, tt.n)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
			if result != tt.expected {
				t.Errorf("Expected result: %v, got: %v", tt.expected, result)
			}
		})
	}
}
