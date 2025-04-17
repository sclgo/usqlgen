package shell

import (
	"testing"

	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {

	t.Run("create tmp dir", func(t *testing.T) {
		cmd := CompileCommand{
			CommandBase: Base(new(GlobalParams)),
			generator: func(input gen.Input) (gen.Result, error) {
				require.DirExists(t, input.WorkingDir)
				return gen.Result{}, nil
			},
			goBin:  "echo",
			Static: true,
		}

		err := cmd.compile("build")
		require.NoError(t, err)
	})

}
