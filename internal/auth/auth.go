package auth

import (
	"context"
	"errors"
	"io"
)

type RequestAuthorizer interface {
	AuthorizeRequest(ctx context.Context, request Request, rw io.ReadWriter) error
}

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrProtoMismatch = errors.New("protocol mismatch")
)

type Request struct {
	ClientAddr string
}
