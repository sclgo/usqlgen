package fi

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Must returns a functions fails the current test with FailNow if err != nil otherwise returns val
// Similar to samber/lo.Must but fails the current test instead of panic
func Must[T any](val T, err error) func(t require.TestingT) T {
	return func(t require.TestingT) T {
		require.NoError(t, err)
		return val
	}
}

// MustF fails the current test if f returns an error. Useful in defer.
func MustF(f func() error, t assert.TestingT) {
	err := f()
	assert.NoError(t, err)
}

// Bind converts a single-parameter function to a no-parameter one by binding the given
// value to the parameter. Useful together with MustF or defer.
func Bind[T any, E any](f func(t T) E, t T) func() E {
	return func() E {
		return f(t)
	}
}
