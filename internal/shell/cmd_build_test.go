package shell

import (
	"bytes"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const main = `
package main

func main() {
}
`

func TestBuild(t *testing.T) {
	t.Run("build to stdout", func(t *testing.T) {
		cmd := BuildCommand{
			CompileCommand: CompileCommand{
				CommandBase: Base(&GlobalParams{}),
				generator:   minimalGoGenerator,
				goBin:       run.FindGo(),
			},
			output: "-",
		}

		var buf bytes.Buffer
		err := cmd.Action(&buf)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		outputTmpFile := filepath.Join(t.TempDir(), "usql")
		err = os.WriteFile(outputTmpFile, buf.Bytes(), 0644)
		require.NoError(t, err)
		result, err := exec.Command("file", outputTmpFile).Output()
		require.NoError(t, err)
		require.Contains(t, string(result), "LSB executable")
		require.Contains(t, string(result), "ELF")
	})
}

func minimalGoGenerator(input gen.Input) error {
	err := run.Go(input.WorkingDir, "mod", "init", "a")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(input.WorkingDir, "main.go"), []byte(main), 0644)
}
