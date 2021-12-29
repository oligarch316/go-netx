package multi

import (
	"context"
	"errors"
	"net"
)

// Dialer TODO.
type Dialer struct{ dialSet }

// Len TODO.
func (d *Dialer) Len() int { return len(d.listeners) }

// SetAddrs TODO.
func (d *Dialer) SetAddrs() []SetAddr {
	res := d.dialSet.SetAddrs()

	// TODO: Order by addressx.Ordering parameter

	return res
}

// Dial TODO.
func (d *Dialer) Dial() (net.Conn, error) {
	return nil, errors.New("not yet implemented")
}

// DialContext TODO.
func (d *Dialer) DialContext(ctx context.Context) (net.Conn, error) {
	return nil, errors.New("not yet implemented")
}
