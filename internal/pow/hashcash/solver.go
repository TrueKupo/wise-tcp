package hashcash

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

type Solver struct {
}

func NewSolver() *Solver {
	return &Solver{}
}

func (s *Solver) Solve(challenge string) (string, error) {
	ch := &Challenge{}
	if err := ch.FromString(challenge); err != nil {
		return "", err
	}
	solution := s.solve(ch)
	if "" == solution {
		return "", fmt.Errorf("solution not found")
	}

	r := &Response{}
	r.FromChallenge(ch, solution)
	return r.String(), nil
}

func (s *Solver) solve(ch *Challenge) string {
	chStr := ch.String()
	bits := ch.Difficulty
	n := bits / 8
	m := bits % 8
	if m > 0 {
		n++
	}

	var result string
	var solution uint32
	sb := make([]byte, 4)
	for {
		binary.LittleEndian.PutUint32(sb, solution)
		result = base64.RawURLEncoding.EncodeToString(sb)
		hash := sha256.Sum256([]byte(chStr + ":" + result))
		isValid, err := verifyBits(hash[:n], bits, n)
		if err != nil {
			continue
		}
		if isValid {
			return result
		}
		solution++
	}
}
