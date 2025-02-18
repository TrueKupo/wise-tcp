package auth

import (
	"context"
	"errors"
	"io"
)

type requestKey struct{}

type Authorizer interface {
	Authorize(ctx context.Context, rw io.ReadWriter) error
}

// RequestAuthorizer has clearer semantics compared to Authorizer
type RequestAuthorizer interface {
	AuthorizeRequest(ctx context.Context, request Request, rw io.ReadWriter) error
}

var (
	ErrRequestNotFound = errors.New("request not found")
	ErrUnauthorized    = errors.New("unauthorized")
)

type Request struct {
	ResourceID string // unused
	ClientAddr string
}

func RequestToContext(ctx context.Context, r Request) context.Context {
	return context.WithValue(ctx, requestKey{}, r)
}

func RequestFromContext(ctx context.Context) (Request, error) {
	v := ctx.Value(requestKey{})
	if v == nil {
		return Request{}, ErrRequestNotFound
	}
	return v.(Request), nil
}
