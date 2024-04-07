package fi_test

import (
	"errors"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/stretchr/testify/require"
	"testing"
)

type testingT func(format string, args ...any)

func (f testingT) Errorf(format string, args ...any) {
	f(format, args...)
}

func TestNoErrorF(t *testing.T) {
	called := false
	var calledStub testingT = func(string, ...any) {
		called = true
	}
	fi.NoErrorF(fi.Bind(errors.New, "test"), calledStub)
	require.Equal(t, true, called)
}
