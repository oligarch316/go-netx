package runner

import (
	"context"
	"fmt"

	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/synctest"
	"github.com/stretchr/testify/assert"
)

// GroupResultsSignal TODO.
type GroupResultsSignal struct{ synctest.SetSignal }

// Value TODO.
func (grs GroupResultsSignal) Value() (res []error) {
	for _, item := range grs.SetSignal.Value() {
		res = append(res, item.(error))
	}
	return
}

// AssertEqualStrings TODO.
func (grs GroupResultsSignal) AssertEqualStrings(t synctest.AssertT, expected ...string) bool {
	t.Helper()
	if len(expected) < 1 {
		return grs.ValueAssertion(assert.Empty)(t)
	}

	elementStringsMatch := func(t synctest.AssertT, actual []interface{}, msgAndArgs ...interface{}) bool {
		actualStrings := make([]string, len(actual))
		for i, item := range actual {
			actualStrings[i] = item.(error).Error()
		}
		return assert.ElementsMatch(t, expected, actualStrings, msgAndArgs...)
	}

	return grs.SetAssertion(elementStringsMatch)(t)
}

// RequireEqualStrings TODO.
func (grs GroupResultsSignal) RequireEqualStrings(t synctest.RequireT, expected ...string) {
	t.Helper()
	if !grs.AssertEqualStrings(t, expected...) {
		t.FailNow()
	}
}

// GroupResults TODO.
type GroupResults struct {
	name string
	c    <-chan error
}

func (gr GroupResults) readAll() []interface{} {
	var res []interface{}
	for err := range gr.c {
		res = append(res, err)
	}
	return res
}

func (gr GroupResults) readN(n int) []interface{} {
	var res []interface{}
	for i := 0; i < n; i++ {
		err, more := <-gr.c
		if !more {
			break
		}
		res = append(res, err)
	}
	return res
}

func (gr GroupResults) String() string { return gr.name }

// All TODO.
func (gr GroupResults) All() GroupResultsSignal {
	name := fmt.Sprintf("%s (all)", gr)
	return GroupResultsSignal{SetSignal: synctest.GoSetSignal(name, gr.readAll)}
}

// Next TODO.
func (gr GroupResults) Next(n int) GroupResultsSignal {
	var (
		name = fmt.Sprintf("%s (all)", gr)
		f    = func() []interface{} { return gr.readN(n) }
	)
	return GroupResultsSignal{SetSignal: synctest.GoSetSignal(name, f)}
}

// Group TODO.
type Group struct {
	name    string
	wrapped *runner.Group
}

// NewGroup TODO.
func NewGroup(name string, items ...runner.Item) Group {
	return WrapGroup(name, runner.NewGroup(items...))
}

// WrapGroup TODO.
func WrapGroup(name string, group *runner.Group) Group {
	return Group{name: name, wrapped: group}
}

func (g Group) String() string { return g.name }

// Append TODO.
func (g Group) Append(items ...runner.Item) { g.wrapped.Append(items...) }

// Close TODO.
func (g Group) Close(ctx context.Context) { g.wrapped.Close(ctx) }

// CloseNow TODO.
func (g Group) CloseNow() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g.Close(ctx)
}

// Run TODO.
func (g Group) Run() GroupResults {
	return GroupResults{
		name: g.name + " <-Results()",
		c:    g.wrapped.Run(),
	}
}
