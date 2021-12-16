package servicex

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi/addrsort"
)

var (
	// DefaultDialKey TODO.
	DefaultDialKey = "localapp"

	// DefaultDialNetworkPriority TODO.
	DefaultDialNetworkPriority = addrsort.ByPriorityAddress(netx.InternalNetwork, "unix", "tcp")
)
