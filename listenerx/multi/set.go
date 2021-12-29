package multi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/oligarch316/go-netx"
)

var (
	gDialSetID uint32

	errInvalidSetHash = errors.New("invalid set hash")
)

// SetAddr TODO.
type SetAddr struct {
	net.Addr
	SetHash
}

// SetHash TODO.
type SetHash interface {
	HashString() string

	id() uint32
	idx() uint32
}

func ParseSetHash(s string) (SetHash, error) {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid hash string '%s': %w", s, err)
	}
	return setHash(val), nil
}

type setHash uint64

func newSetHash(setID, setIdx uint32) setHash { return setHash(uint64(setID)<<32 | uint64(setIdx)) }

func (sh setHash) HashString() string { return strconv.FormatUint(uint64(sh), 10) }
func (sh setHash) id() uint32         { return uint32(sh >> 32) }
func (sh setHash) idx() uint32        { return uint32(sh & 0x7FFFFFFF) }

type dialSet struct {
	id        uint32
	listeners []netx.Listener
}

func newDialSet() dialSet { return dialSet{id: atomic.AddUint32(&gDialSetID, 1)} }

func (ds dialSet) lookup(hash SetHash) (netx.Listener, error) {
	hID, hIdx := hash.id(), hash.idx()

	switch {
	case hID != ds.id:
		return nil, fmt.Errorf("%w: hash id '%d' does not match set id '%d", errInvalidSetHash, hID, ds.id)
	case hIdx >= uint32(len(ds.listeners)):
		return nil, fmt.Errorf("%w: hash index '%d' out of bounds", errInvalidSetHash, hIdx)
	}

	return ds.listeners[hIdx], nil
}

func (ds dialSet) SetAddrs() []SetAddr {
	res := make([]SetAddr, 0, len(ds.listeners))

	for i, l := range ds.listeners {
		res = append(res, SetAddr{
			Addr:    l.Addr(),
			SetHash: newSetHash(ds.id, uint32(i)),
		})
	}

	return res
}

func (ds dialSet) DialHash(hash SetHash) (net.Conn, error) {
	l, err := ds.lookup(hash)
	if err != nil {
		return nil, err
	}

	return l.Dial()
}

func (ds dialSet) DialContextHash(ctx context.Context, hash SetHash) (net.Conn, error) {
	l, err := ds.lookup(hash)
	if err != nil {
		return nil, err
	}

	return l.DialContext(ctx)
}
