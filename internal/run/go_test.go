package run_test

import (
	"testing"

	"github.com/sclgo/usqlgen/internal/run"
	"github.com/stretchr/testify/require"
)

func TestGoBin(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		err := run.GoBin(".", nil, "test")
		require.Error(t, err)
	})
}
