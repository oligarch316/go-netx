package servicex

import (
	"github.com/oligarch316/go-netx/addressx"
	"github.com/oligarch316/go-netx/listenerx"
)

var (
	// DefaultDialKey TODO.
	DefaultDialKey = "localapp"

	// DefaultDialNetworkPriority TODO.
	DefaultDialNetworkPriority = addressx.ByPriorityAddress(listenerx.InternalNetwork, "unix", "tcp")
)
