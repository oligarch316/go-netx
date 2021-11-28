package multi_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/oligarch316/go-netx/synctest"
	multitest "github.com/oligarch316/go-netx/synctest/multi"
)

const mlSize = 3

func setupListener(name string, size int) (multitest.Listener, []mockListener) {
	var (
		ml    = multitest.NewListener(name)
		mocks = make([]mockListener, size)
	)

	for i := 0; i < size; i++ {
		mocks[i] = newMockListener(fmt.Sprintf("mock listener %d", i))
		ml.Append(mocks[i])
	}

	return ml, mocks
}

func TestConcurrentListener(t *testing.T) {
	ml, mocks := setupListener("multi listener", mlSize)

	type task struct {
		name string
		f    func()
	}

	var tasks []task

	for _, addr := range ml.Addrs() {
		tasks = append(tasks, task{
			name: fmt.Sprintf("%s dial (%s)", ml, addr),
			f:    func() { ml.Dial(addr) },
		})
	}

	for _, mock := range mocks {
		name := fmt.Sprintf("%s send", mock)
		f := func() { mock.sendConn(mockConn{name: name + " conn"}) }
		tasks = append(tasks, task{name: name, f: f})
	}

	rand.Shuffle(len(tasks), func(i, j int) {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	})

	// Duh need to use the ml.Runners() in here somewhere idiot

	acceptSig := ml.Accept(mlSize * 2)

	taskSigs := make(synctest.SignalList, len(tasks))
	for i, task := range tasks {
		taskSigs[i] = synctest.GoSignal(task.name, task.f)
	}

	taskSigs.RequireState(t, synctest.Complete.After(1*time.Second))
	acceptSig.RequireState(t, synctest.Complete)

	acceptSig.AssertConns(t, "TODO")
	acceptSig.AssertError(t, nil)
}
