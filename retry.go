package netx

import "time"

const (
	retryDelayMin = 5 * time.Millisecond
	retryDelayMax = 1 * time.Second
)

type retryDelay struct{ time.Duration }

func newRetryDelay() *retryDelay {
	res := new(retryDelay)
	res.Reset()
	return res
}

func (rd *retryDelay) Reset() { rd.Duration = retryDelayMin }

func (rd *retryDelay) Sleep() <-chan time.Time {
	res := time.After(rd.Duration)

	if rd.Duration < retryDelayMax {
		rd.Duration *= 2

		if rd.Duration > retryDelayMax {
			rd.Duration = retryDelayMax
		}
	}

	return res
}
