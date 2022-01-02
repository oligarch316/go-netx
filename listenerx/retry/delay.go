package retry

import "time"

// Delay TODO.
type Delay struct {
	attempt   int
	delayFunc DelayFunc
}

// NewDelay TODO.
func NewDelay(f DelayFunc) *Delay { return &Delay{delayFunc: f} }

// Reset TODO.
func (d *Delay) Reset() { d.attempt = 0 }

// Next TODO.
func (d *Delay) Next() (int, time.Duration) {
	d.attempt++
	return d.attempt, d.delayFunc(d.attempt)
}
