package multi

import (
	"fmt"
	"net"
	"strconv"
)

// Addr TODO.
type Addr struct {
	net.Addr
	Hash
}

// Hash TODO.
type Hash interface {
	HashString() string

	id() uint32
	idx() uint32
}

// ParseHash TODO.
func ParseHash(s string) (Hash, error) {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid hash string '%s': %w", s, err)
	}
	return hash(val), nil
}

type hash uint64

func newHash(setID, setIdx uint32) hash { return hash(uint64(setID)<<32 | uint64(setIdx)) }

func (h hash) HashString() string { return strconv.FormatUint(uint64(h), 10) }

func (h hash) id() uint32 { return uint32(h >> 32) }

func (h hash) idx() uint32 { return uint32(h & 0x7FFFFFFF) }
