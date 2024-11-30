package shell

import (
	"bytes"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

const main = `
package main

func main() {
}
`

func TestBuild(t *testing.T) {
	t.Run("stdout", func(t *testing.T) {
		cmd := BuildCommand{
			CompileCommand: CompileCommand{
				CommandBase: Base(&GlobalParams{}),
				generator: func(input gen.Input) error {
					err := run.Go(input.WorkingDir, "mod", "init", "a")
					if err != nil {
						return err
					}
					return os.WriteFile(filepath.Join(input.WorkingDir, "main.go"), []byte(main), 0644)
				},
				goBin: run.FindGo(),
			},
			output: "-",
		}

		var buf bytes.Buffer
		err := cmd.Action(&buf)
		require.NoError(t, err)
		require.NotEmpty(t, buf)
	})
}
