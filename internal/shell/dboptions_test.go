package shell

import (
	"testing"

	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/stretchr/testify/require"
)

func TestApplyOptionsFromNames(t *testing.T) {
	t.Run("cse insensitive", func(t *testing.T) {
		genInput := gen.Input{}
		err := applyOptionsFromNames([]string{"IncludeSemicoloN"}, &genInput)
		require.NoError(t, err)
		require.True(t, genInput.IncludeSemicolon)
	})
	t.Run("bad opt, good opt", func(t *testing.T) {
		genInput := gen.Input{}
		err := applyOptionsFromNames([]string{"includeSemicolon", "foobar"}, &genInput)
		require.ErrorContains(t, err, "foobar")
		require.False(t, genInput.IncludeSemicolon)
	})
}
