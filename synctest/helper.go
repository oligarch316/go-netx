package synctest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertT TODO.
type AssertT interface {
	assert.TestingT
	Helper()
}

// RequireT TODO.
type RequireT interface {
	require.TestingT
	Helper()
}
