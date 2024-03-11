package fi

import (
	"github.com/stretchr/testify/require"
)

func Must[T any](val T, err error) func(t require.TestingT) T {
	return func(t require.TestingT) T {
		require.NoError(t, err)
		return val
	}
}
