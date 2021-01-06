package multi

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/oligarch316/go-netx"
)

var gSetID uint32

// Set TODO.
type Set struct {
	id        uint32
	listeners []netx.Listener
}

func newSet() Set { return Set{id: atomic.AddUint32(&gSetID, 1)} }

func (s Set) lookup(h Hash) (netx.Listener, error) {
	hID, hIdx := h.id(), h.idx()

	switch {
	case hID != s.id:
		return nil, fmt.Errorf("invalid hash id '%d': does not match set id '%d'", hID, s.id)
	case hIdx >= uint32(len(s.listeners)):
		return nil, fmt.Errorf("invalid hash index '%d': out of bounds", hIdx)
	}

	return s.listeners[hIdx], nil
}

// Append TODO.
func (s *Set) Append(ls ...netx.Listener) { s.listeners = append(s.listeners, ls...) }

// ID TODO.
func (s Set) ID() uint32 { return s.id }

// Len TODO.
func (s Set) Len() int { return len(s.listeners) }

// Addrs TODO.
func (s Set) Addrs() (res []Addr) {
	for idx, l := range s.listeners {
		res = append(res, Addr{
			Addr: l.Addr(),
			Hash: newHash(s.id, uint32(idx)),
		})
	}

	return
}

// Dial TODO.
func (s Set) Dial(h Hash) (net.Conn, error) {
	l, err := s.lookup(h)
	if err != nil {
		return nil, err
	}

	return l.Dial()
}

// DialContext TODO.
func (s Set) DialContext(ctx context.Context, h Hash) (net.Conn, error) {
	l, err := s.lookup(h)
	if err != nil {
		return nil, err
	}

	return l.DialContext(ctx)
}
