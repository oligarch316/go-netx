package httpx

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/oligarch316/go-netx"
)

var dialLocalHostKey = fmt.Sprintf("_%s_:0", ID)

type dialContextFunc func(context.Context, string, string) (net.Conn, error)

type dialHooks struct {
	dial        func(string, string) (net.Conn, error)
	dialContext dialContextFunc
}

func (dh dialHooks) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var (
		res net.Conn
		err error
	)

	switch {
	case dh.dialContext != nil:
		res, err = dh.dialContext(ctx, network, addr)
	case dh.dial != nil:
		res, err = dh.dial(network, addr)
	default:
		res, err = (&net.Dialer{}).DialContext(ctx, network, addr)
	}

	if res == nil && err == nil {
		err = errors.New("httpx: transport dial hook returned (nil, nil)")
	}

	return res, err
}

func wrapDialContext(dialer netx.Dialer, f dialContextFunc) dialContextFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr == dialLocalHostKey {
			return dialer.DialContext(ctx)
		}
		return f(ctx, network, addr)
	}
}
