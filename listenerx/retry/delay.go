package retry

import "time"

type Delay struct {
	attempt   int
	delayFunc DelayFunc
}

func NewDelay(f DelayFunc) *Delay { return &Delay{delayFunc: f} }

func (d *Delay) Reset() { d.attempt = 0 }

func (d *Delay) Next() (int, time.Duration) {
	d.attempt++
	return d.attempt, d.delayFunc(d.attempt)
}
