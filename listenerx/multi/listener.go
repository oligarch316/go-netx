package multi

import (
	"fmt"

	"github.com/oligarch316/go-netx"
)

// Listener TODO.
type Listener struct {
	*Dialer
	mergeListener
}

// NewListener TODO.
func NewListener(ls ...netx.Listener) *Listener {
	res := &Listener{
		Dialer:        newDialer(),
		mergeListener: newMergeListener(),
	}

	res.Append(ls...)
	return res
}

// Append TODO.
func (l *Listener) Append(ls ...netx.Listener) { l.set.listeners = append(l.set.listeners, ls...) }

// Runners TODO.
func (l *Listener) Runners() []*mergeRunner {
	res := make([]*mergeRunner, l.Len())

	for i, item := range l.set.listeners {
		res[i] = newMergeRunner(item, l.mergeListenerChannels)
	}

	return res
}

func (l *Listener) String() string { return fmt.Sprintf("multi listener %d", l.SetID()) }
