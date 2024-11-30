package shell

import (
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCompile(t *testing.T) {

	t.Run("create tmp dir", func(t *testing.T) {
		cmd := CompileCommand{
			CommandBase: Base(new(GlobalParams)),
			generator: func(input gen.Input) error {
				require.DirExists(t, input.WorkingDir)
				return nil
			},
			goBin: "echo",
		}

		err := cmd.compile("build")
		require.NoError(t, err)
	})

}
