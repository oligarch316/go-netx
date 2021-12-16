package serverx

import (
	"context"
	"fmt"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/runner"
)

// RunnerAction TODO.
type RunnerAction string

const (
	// RunnerActionRun TODO.
	RunnerActionRun RunnerAction = "Run"

	// RunnerActionClose TODO.
	RunnerActionClose RunnerAction = "Close"
)

// RunnerInfo TODO.
type RunnerInfo struct {
	Name      string
	ServiceID netx.ServiceID
}

// RunnerError TODO.
type RunnerError struct {
	error
	RunnerInfo
	Action RunnerAction
}

func (re RunnerError) Unwrap() error { return re.error }

func (re RunnerError) Error() string {
	return fmt.Sprintf("%s %s '%s': %s", re.ServiceID, re.Name, re.Action, re.error)
}

type serverRunner struct {
	runner.Item
	RunnerInfo
}

func newServerRunner(svcID netx.ServiceID, name string, rnr runner.Item) *serverRunner {
	return &serverRunner{
		Item: rnr,
		RunnerInfo: RunnerInfo{
			Name:      name,
			ServiceID: svcID,
		},
	}
}

func (sr serverRunner) Run() error {
	if err := sr.Item.Run(); err != nil {
		return RunnerError{
			error:      err,
			RunnerInfo: sr.RunnerInfo,
			Action:     RunnerActionRun,
		}
	}

	return nil
}

func (sr serverRunner) Close(ctx context.Context) error {
	if err := sr.Item.Close(ctx); err != nil {
		return RunnerError{
			error:      err,
			RunnerInfo: sr.RunnerInfo,
			Action:     RunnerActionClose,
		}
	}

	return nil
}
