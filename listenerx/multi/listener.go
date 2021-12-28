package multi

import (
	"fmt"

	"github.com/oligarch316/go-netx"
)

// Listener TODO.
type Listener struct {
	mergeListener
	Set
}

// NewListener TODO.
func NewListener(ls ...netx.Listener) *Listener {
	res := &Listener{
		mergeListener: newMergeListener(),
		Set:           newSet(),
	}

	res.Append(ls...)
	return res
}

// Runners TODO.
func (l *Listener) Runners() []*mergeRunner {
	res := make([]*mergeRunner, l.Len())

	for i, item := range l.listeners {
		res[i] = newMergeRunner(item, l.mergeListenerChannels)
	}

	return res
}

func (l *Listener) String() string {
	return fmt.Sprintf("multi listener %d", l.ID())
}
