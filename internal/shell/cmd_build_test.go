package shell

import (
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuild(t *testing.T) {
	t.Run("stdout", func(t *testing.T) {
		cmd := BuildCommand{
			CompileCommand: CompileCommand{
				CommandBase: Base(new(GlobalParams)),
				generator: func(input gen.Input) error {
					return nil
				},
				goBin: "echo",
			},
			output: "-",
		}

		err := cmd.Action(nil)
		require.NoError(t, err)
		// TODO complete test
	})
}
