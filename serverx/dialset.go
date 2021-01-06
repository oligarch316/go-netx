package serverx

import (
	"context"
	"errors"
	"net"

	"github.com/oligarch316/go-netx/multi"
	"github.com/oligarch316/go-netx/multi/addrsort"
)

var errUnknownDialFailure = errors.New("serverx: unknown dial failure")

// DialSet TODO.
type DialSet struct{ mSet multi.Set }

// Dial TODO.
func (ds DialSet) Dial(hs ...multi.Hash) (net.Conn, error) {
	var firstErr error

	for _, h := range hs {
		res, err := ds.mSet.Dial(h)
		switch {
		case err == nil:
			return res, nil
		case firstErr == nil:
			firstErr = err
		}
	}

	if firstErr == nil {
		return nil, errUnknownDialFailure
	}

	return nil, firstErr
}

// DialContext TODO.
func (ds DialSet) DialContext(ctx context.Context, hs ...multi.Hash) (net.Conn, error) {
	var firstErr error

	for _, h := range hs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			res, err := ds.mSet.DialContext(ctx, h)
			switch {
			case err == nil:
				return res, nil
			case firstErr == nil:
				firstErr = err
			}
		}
	}

	if firstErr == nil {
		return nil, errUnknownDialFailure
	}

	return nil, firstErr
}

// Resolve TODO.
// TODO: is returning []multi.Hash instead of []multi.Addr really the best idea
// here. What about callers wanting to log/error with net.Addr info when issues occur?
func (ds DialSet) Resolve(cmps ...addrsort.Comparer) []multi.Hash {
	var (
		addrs = ds.mSet.Addrs()
		res   = make([]multi.Hash, len(addrs))
	)

	addrsort.Stable(addrs, cmps...)

	for i, addr := range addrs {
		res[i] = addr
	}

	return res
}
