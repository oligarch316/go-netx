package grpcx

import "fmt"

// Error TODO.
type Error struct {
	Component string
	err error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s %s: %s", namespace, e.Component, e.err.Error())
}

// Unwrap TODO.
func (e Error) Unwrap() error {
	return e.err
}
