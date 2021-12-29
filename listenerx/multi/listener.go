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
		Dialer:        &Dialer{dialSet: newDialSet()},
		mergeListener: newMergeListener(),
	}

	res.Append(ls...)
	return res
}

// Append TODO.
func (l *Listener) Append(ls ...netx.Listener) { l.listeners = append(l.listeners, ls...) }

// Runners TODO.
func (l *Listener) Runners() []*mergeRunner {
	res := make([]*mergeRunner, len(l.listeners))

	for i, item := range l.listeners {
		res[i] = newMergeRunner(item, l.mergeListenerChannels)
	}

	return res
}

func (l *Listener) String() string { return fmt.Sprintf("multi listener %d", l.id) }
