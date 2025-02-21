package hashcash

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Payload struct {
	Version    int
	Difficulty int
	ExpiresAt  time.Time
	Subject    string
	Nonce      string
	Alg        string
}

func (p *Payload) String(extraParts ...string) string {
	parts := append([]string{
		strconv.Itoa(p.Version),
		strconv.Itoa(p.Difficulty),
		strconv.FormatInt(p.ExpiresAt.Unix(), 10),
		p.Subject,
		p.Nonce,
		p.Alg,
	}, extraParts...)
	return strings.Join(parts, ":")
}

func (p *Payload) FromString(parts []string) error {
	if len(parts) < 6 {
		return fmt.Errorf("invalid payload string")
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil || version != 1 {
		return fmt.Errorf("invalid version")
	}
	p.Version = version

	bits, err := strconv.Atoi(parts[1])
	if err != nil || bits <= 0 || bits > maxDifficulty {
		return fmt.Errorf("invalid difficulty")
	}
	p.Difficulty = bits

	exp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || exp < 0 {
		return fmt.Errorf("invalid expiration")
	}
	expAt := time.Unix(exp, 0).UTC()
	if expAt.Before(time.Now().UTC()) {
		return fmt.Errorf("expiration in the past")
	}
	p.ExpiresAt = expAt

	if strings.TrimSpace(parts[3]) == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	p.Subject = parts[3]

	if strings.TrimSpace(parts[4]) == "" {
		return fmt.Errorf("nonce cannot be empty")
	}
	p.Nonce = parts[4]

	if strings.TrimSpace(parts[5]) == "" {
		return fmt.Errorf("algorithm cannot be empty")
	}
	p.Alg = parts[5]

	return nil
}

func (p *Payload) Fingerprint() (string, error) {
	f, err := getFingerprint(p.Alg, p.Subject, p.Nonce, p.ExpiresAt, p.Difficulty)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(f)), nil
}

func getFingerprint(alg, subject, nonce string, at time.Time, difficulty int) (string, error) {
	alg = strings.TrimSpace(alg)
	subject = strings.TrimSpace(subject)
	nonce = strings.TrimSpace(nonce)
	if alg == "" || subject == "" || nonce == "" {
		return "", fmt.Errorf("invalid input: algorithm, subject, nonce cannot be empty")
	}
	if at.Before(time.Now().UTC()) {
		return "", fmt.Errorf("invalid input: timestamp cannot be in the past")
	}
	if difficulty <= 0 {
		return "", fmt.Errorf("invalid input: difficulty must be positive")
	}

	return alg + ":" +
		subject + ":" +
		nonce + ":" +
		strconv.FormatInt(at.UTC().Unix(), 10) + ":" +
		strconv.Itoa(difficulty), nil
}
