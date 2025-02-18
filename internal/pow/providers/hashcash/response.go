package hashcash

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type Response struct {
	Payload
	Solution string
	verifier VerifyFunc
}

var ErrInvalidSolution = fmt.Errorf("invalid solution")

type ResponseOption func(*Response)

func WithVerifier(v VerifyFunc) ResponseOption {
	return func(r *Response) {
		r.verifier = v
	}
}

func WithPayload(p Payload) ResponseOption {
	return func(r *Response) {
		r.Payload = p
	}
}

func NewResponse(solution string, opts ...ResponseOption) *Response {
	r := &Response{
		Solution: solution,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func ResponseFromChallenge(challenge string, solution string, opts ...ResponseOption) (*Response, error) {
	parsed, err := ParseChallenge(challenge)
	if err != nil {
		return nil, err
	}

	opts = append(opts, WithPayload(parsed.Payload))
	response := NewResponse(solution, opts...)

	return response, nil
}

func ParseChallenge(challenge string) (*Challenge, error) {
	result := &Challenge{}
	if err := result.FromString(challenge); err != nil {
		return nil, fmt.Errorf("failed to parse challenge: %v", err)
	}

	return result, nil
}

func (r *Response) FromString(str string) error {
	parts := strings.Split(str, ":")
	if len(parts) != 7 {
		return fmt.Errorf("invalid response string")
	}

	if err := r.Payload.FromString(parts[:6]); err != nil {
		return err
	}

	r.Solution = parts[6]
	return nil
}

func (r *Response) String() string {
	return r.Payload.String(r.Solution)
}

func (r *Response) FromChallenge(ch *Challenge, solution string) {
	r.Payload = ch.Payload
	r.Solution = solution
}

func (r *Response) Verify() error {
	if r.verifier == nil {
		r.verifier = verifyBits
	}

	bits := r.Difficulty
	hash := sha256.Sum256([]byte(r.String()))
	n := bits / 8
	m := bits % 8
	if m > 0 {
		n++
	}

	valid, err := r.verifier(hash[:n], bits, n)
	if err != nil {
		return fmt.Errorf("verification failed: %v", err)
	}
	if !valid {
		return ErrInvalidSolution
	}
	return nil
}
