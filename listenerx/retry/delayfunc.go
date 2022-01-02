package retry

import (
	"math"
	"time"
)

// DelayFunc TODO.
type DelayFunc func(attempt int) (delay time.Duration)

// DelayFuncConstant TODO.
func DelayFuncConstant(duration time.Duration) DelayFunc {
	return func(int) time.Duration { return duration }
}

// DelayFuncMultiplicative TODO.
func DelayFuncMultiplicative(min, max time.Duration, factor float64) DelayFunc {
	maxF := float64(max)

	return func(attempt int) time.Duration {
		resF := float64(min) * factor * float64(attempt)
		return time.Duration(math.Min(resF, maxF))
	}
}

// DelayFuncExponential TODO.
func DelayFuncExponential(min, max time.Duration, factor float64) DelayFunc {
	maxF := float64(max)

	return func(attempt int) time.Duration {
		resF := float64(min) * math.Pow(factor, float64(attempt))
		return time.Duration(math.Min(resF, maxF))
	}
}
