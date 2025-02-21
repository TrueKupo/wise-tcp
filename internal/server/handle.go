package server

import (
	"context"
	"errors"
	"net"

	"wise-tcp/internal/auth"
	"wise-tcp/pkg/log"
)

type connHandler struct {
	throttle   *Throttle
	auth       auth.RequestAuthorizer
	reqHandler RequestHandler
}

func (h *connHandler) Handle(ctx context.Context, conn net.Conn) {
	if err := h.throttle.Acquire(ctx, conn); err != nil {
		if errors.Is(err, ErrConnRejected) || errors.Is(err, ErrConnDropped) {
			log.Warnf("Connection throttled: %v", err)
			return
		}
		log.Errorf("Acquire error: %v", err)
		return
	}
	defer h.throttle.Release()
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Errorf("Close connection: %v", err)
		}
	}(conn)

	dl, ok := ctx.Deadline()
	if !ok {
		log.Warnf("Context deadline not set")
		return
	}
	if err := conn.SetDeadline(dl); err != nil {
		log.Errorf("Failed to set connection deadline: %v", err)
		return
	}

	if h.auth != nil {
		req := auth.Request{
			ClientAddr: conn.RemoteAddr().String(),
		}
		if err := h.auth.AuthorizeRequest(ctx, req, conn); err != nil {
			if errors.Is(err, auth.ErrUnauthorized) {
				log.Warn("Unauthorized request")
			} else {
				log.Errorf("Authorize error: %v", err)
			}
			return
		}
	}

	if err := h.reqHandler.Handle(ctx, conn); err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			log.Warn("Connection timed out during processing")
		} else {
			log.Errorf("Handler failed to process request: %v", err)
		}
	}
}
