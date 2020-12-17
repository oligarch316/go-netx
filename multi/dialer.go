package multi

import (
    "context"
    "errors"
    "fmt"
    "net"
    "sort"
    "strconv"
    "sync/atomic"

    "github.com/oligarch316/go-netx"
)

var (
    // ErrInvalidHash TODO.
    ErrInvalidHash = errors.New("invalid multi hash")

    gSetID uint32
)

// Dialer TODO.
type Dialer interface {
    Addrs() []Addr
    Dial(Hash) (net.Conn, error)
    DialContext(context.Context, Hash) (net.Conn, error)
}

// Addr TODO.
type Addr struct {
    net.Addr
    Hash
}

// AddrSorter TODO.
type AddrSorter []netx.AddrComparer

// Sort TODO.
func (as AddrSorter) Sort(addrs []Addr) {
    sort.Slice(addrs, func(i, j int) bool { return netx.AddrSorter(as).Less(addrs[i], addrs[j]) })
}

// Stable TODO.
func (as AddrSorter) Stable(addrs []Addr) {
    sort.SliceStable(addrs, func(i, j int) bool { return netx.AddrSorter(as).Less(addrs[i], addrs[j]) })
}

// Hash TODO.
type Hash interface {
    Format() string

    id() uint32
    idx() uint32
}

// ParseHash TODO.
func ParseHash(s string) (Hash, error) {
    val, err := strconv.ParseUint(s, 10, 64)
    if err != nil {
        return nil, fmt.Errorf("invalid hash: %w", err)
    }
    return hash(val), nil
}

type hash uint64

func newHash(setID, setIdx uint32) hash { return hash(uint64(setID) << 32 | uint64(setIdx)) }

func (h hash) Format() string { return strconv.FormatUint(uint64(h), 10) }

func (h hash) id() uint32 { return uint32(h >> 32) }

func (h hash) idx() uint32 { return uint32(h & 0x7FFFFFFF) }

type set struct {
	id        uint32
	listeners []netx.Listener
}

func newSet() set { return set{id: atomic.AddUint32(&gSetID, 1)} }

func (s set) lookup(h Hash) (netx.Listener, error) {
    hID, hIdx := h.id(), h.idx()

    switch {
    case hID != s.id:
        return nil, fmt.Errorf("%w: hash id '%d' does not match listener id '%d'", ErrInvalidHash, hID, s.id)
    case hIdx >= uint32(len(s.listeners)):
        return nil, fmt.Errorf("%w: hash index '%d' out of bounds", ErrInvalidHash, hIdx)
    }

    return s.listeners[hIdx], nil
}

func (s *set) Append(ls ...netx.Listener) { s.listeners = append(s.listeners, ls...) }

func (s set) ID() uint32 { return s.id }

func (s set) Len() int { return len(s.listeners) }

func (s set) Addrs() (res []Addr) {
    for idx, l := range s.listeners {
        res = append(res, Addr{
            Addr: l.Addr(),
            Hash: newHash(s.id, uint32(idx)),
        })
    }

    return
}

func (s set) Dial(h Hash) (net.Conn, error) {
	l, err := s.lookup(h)
	if err != nil {
		return nil, err
	}

	return l.Dial()
}

func (s set) DialContext(ctx context.Context, h Hash) (net.Conn, error) {
	l, err := s.lookup(h)
	if err != nil {
		return nil, err
	}

	return l.DialContext(ctx)
}
