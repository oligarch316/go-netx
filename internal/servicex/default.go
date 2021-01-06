package servicex

import "github.com/oligarch316/go-netx"

// DefaultKey TODO.
const DefaultKey = "localapp"

// DefaultNetworkPriority TODO.
var DefaultNetworkPriority = []string{
	netx.NetworkInternal,
	"unix",
	"tcp",
}
