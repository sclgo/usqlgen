package shell

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	t.Run("happy path", func(t *testing.T) {
		cmd := GenerateCommand{
			CompileCommand: minimalCompileCommand(),
			output:         tmpDir,
		}
		err := cmd.Action(nil)
		require.NoError(t, err)

		require.FileExists(t, filepath.Join(tmpDir, "go.mod"))
	})

}
