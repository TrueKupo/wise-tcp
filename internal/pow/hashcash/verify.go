package hashcash

import (
	"fmt"
)

type VerifyFunc func(hash []byte, bits, n int) (bool, error)

const bitsInByte = 8

func verifyBits(hash []byte, bits, n int) (bool, error) {
	if bits < 0 || bits > len(hash)*bitsInByte {
		return false,
			fmt.Errorf("invalid input: bits must be between 0 and %d (hash length: %d bytes)",
				len(hash)*8, len(hash))
	}

	fullBytes := bits / bitsInByte
	remainderBits := bits % bitsInByte

	n = min(n, len(hash))

	for i := 0; i < min(fullBytes, n); i++ {
		if hash[i] != 0 {
			return false, nil
		}
	}

	if fullBytes < n && remainderBits > 0 {
		pad := bitsInByte - remainderBits
		if hash[fullBytes]>>pad != 0 {
			return false, nil
		}
	}

	return true, nil
}
