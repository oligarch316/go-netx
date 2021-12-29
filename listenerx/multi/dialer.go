package multi

import (
	"context"
	"errors"
	"net"
)

// Dialer TODO.
type Dialer struct{ set dialSet }

func newDialer() *Dialer { return &Dialer{set: newDialSet()} }

// SetID TODO.
func (d *Dialer) SetID() uint32 { return d.set.id }

// Len TODO.
func (d *Dialer) Len() int { return len(d.set.listeners) }

// Resolve TODO.
func (d *Dialer) Resolve() []SetAddr {
	res := d.set.Addrs()

	// TODO: Order by addressx.Ordering parameter

	return res
}

// DialHash TODO.
func (d *Dialer) DialHash(hash SetHash) (net.Conn, error) {
	return d.set.Dial(hash)
}

// DialContextHash TODO.
func (d *Dialer) DialContextHash(ctx context.Context, hash SetHash) (net.Conn, error) {
	return d.set.DialContext(ctx, hash)
}

// Dial TODO.
func (d *Dialer) Dial() (net.Conn, error) {
	return nil, errors.New("not yet implemented")
}

// DialContext TODO.
func (d *Dialer) DialContext(ctx context.Context) (net.Conn, error) {
	return nil, errors.New("not yet implemented")
}
