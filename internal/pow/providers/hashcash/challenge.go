package hashcash

import (
	"strings"
)

const maxDifficulty = 52

type Challenge struct {
	Payload
}

func (c *Challenge) FromString(str string) error {
	parts := strings.Split(str, ":")
	if err := c.Payload.FromString(parts); err != nil {
		return err
	}
	return nil
}
