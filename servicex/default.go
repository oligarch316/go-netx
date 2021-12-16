package servicex

import (
	"github.com/oligarch316/go-netx/listenerx"
	"github.com/oligarch316/go-netx/listenerx/multi/addrsort"
)

var (
	// DefaultDialKey TODO.
	DefaultDialKey = "localapp"

	// DefaultDialNetworkPriority TODO.
	DefaultDialNetworkPriority = addrsort.ByPriorityAddress(listenerx.InternalNetwork, "unix", "tcp")
)
