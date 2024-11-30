package run_test

import (
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGoBin(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		err := run.GoBin(".", "test")
		require.Error(t, err)
	})
}
