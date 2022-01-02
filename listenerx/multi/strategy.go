package multi

import (
	"context"
	"errors"
	"net"
)

var (
	errEmptyAddressList   = errors.New("empty address list")
	errUnknownDialFailure = errors.New("unknown dial failure")
)

// DialHashFunc TODO.
type DialHashFunc func(context.Context, SetHash) (net.Conn, error)

// DialStrategy TODO.
type DialStrategy func(ctx context.Context, addrs []SetAddr, dialHash DialHashFunc) (net.Conn, error)

// DialStrategyFirstOnly TODO.
func DialStrategyFirstOnly(ctx context.Context, addrs []SetAddr, dialHash DialHashFunc) (net.Conn, error) {
	if (len(addrs)) < 1 {
		return nil, errEmptyAddressList
	}

	return dialHash(ctx, addrs[0])
}

// DialStrategyIterative TODO.
func DialStrategyIterative(ctx context.Context, addrs []SetAddr, dialHash DialHashFunc) (net.Conn, error) {
	if (len(addrs)) < 1 {
		return nil, errEmptyAddressList
	}

	var firstErr error

	for _, addr := range addrs {
		res, err := dialHash(ctx, addr)
		switch {
		case err == nil:
			return res, nil
		case firstErr == nil:
			firstErr = err
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return nil, errUnknownDialFailure
}
