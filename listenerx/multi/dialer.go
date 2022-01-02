package multi

import (
	"context"
	"net"
	"sort"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/addressx"
)

// DialerParams TODO.
type DialerParams struct {
	AddressOrdering addressx.Ordering
	Strategy        DialStrategy
}

// Dialer TODO.
type Dialer struct {
	params DialerParams
	set    *dialSet
}

func newDialer(params DialerParams, ls []netx.Listener) *Dialer {
	return &Dialer{params: params, set: newDialSet(ls)}
}

// Len TODO.
func (d *Dialer) Len() int { return len(d.set.listeners) }

// Resolve TODO.
func (d *Dialer) Resolve() []SetAddr {
	res := d.set.Addrs()

	sort.SliceStable(res, func(i, j int) bool {
		return d.params.AddressOrdering.Less(res[i], res[j])
	})

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
	return d.DialContext(context.Background())
}

// DialContext TODO.
func (d *Dialer) DialContext(ctx context.Context) (net.Conn, error) {
	return d.params.Strategy(ctx, d.Resolve(), d.DialContextHash)
}
