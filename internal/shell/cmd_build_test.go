package shell

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/stretchr/testify/require"
)

const main = `
package main

func main() {
}
`

func TestBuild(t *testing.T) {
	tmpDir := t.TempDir()
	t.Run("build to stdout", func(t *testing.T) {
		cmd := BuildCommand{
			CompileCommand: minimalCompileCommand(),
			output:         "-",
		}

		var buf bytes.Buffer
		err := cmd.Action(&buf)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		outputTmpFile := filepath.Join(tmpDir, "usql")
		err = os.WriteFile(outputTmpFile, buf.Bytes(), 0644)
		require.NoError(t, err)
		checkExecutable(t, outputTmpFile)
	})

	t.Run("default", func(t *testing.T) {
		cmd := BuildCommand{
			CompileCommand: minimalCompileCommand(),
			output:         ".",
		}
		currentWorkDir := fi.NoError(os.Getwd()).Require(t)
		defer fi.NoErrorF(fi.Bind(os.Chdir, currentWorkDir), t)
		require.NoError(t, os.Chdir(tmpDir))
		err := cmd.Action(nil)
		require.NoError(t, err)

		outputTmpFile := filepath.Join(tmpDir, "usql")
		checkExecutable(t, outputTmpFile)

	})

	t.Run("version", func(t *testing.T) {
		cmd := BuildCommand{
			CompileCommand: minimalCompileCommand(),
			output:         "-",
		}
		testVersion := "foobar"
		cmd.CompileCommand.generator = func(input gen.Input) (gen.Result, error) {
			res, err := minimalGoGenerator(input)
			res.DownloadedUsqlVersion = testVersion
			return res, err
		}

		var buf bytes.Buffer
		err := cmd.Action(&buf)
		require.NoError(t, err)
		require.Contains(t, buf.String(), testVersion+"_usqlgen")
	})

	t.Run("full build", func(t *testing.T) {
		fi.SkipLongTest(t)

		cmd := NewCommands(nil).BuildCmd
		require.NoError(t, cmd.Imports.Set("github.com/sclgo/impala-go"))

		currentWorkDir := fi.NoError(os.Getwd()).Require(t)
		defer fi.NoErrorF(fi.Bind(os.Chdir, currentWorkDir), t)
		require.NoError(t, os.Chdir(tmpDir))
		err := cmd.Action(nil)
		require.NoError(t, err)

		outputTmpFile := filepath.Join(tmpDir, "usql")
		checkExecutable(t, outputTmpFile)
	})
}

func checkExecutable(t *testing.T, path string) {
	_, err := os.Stat(path)
	require.NoError(t, err)
	result, err := exec.Command("file", path).Output()
	require.NoError(t, err)
	require.Contains(t, string(result), "LSB executable")
	require.Contains(t, string(result), "ELF")
}

func minimalCompileCommand() CompileCommand {
	return CompileCommand{
		CommandBase: Base(&GlobalParams{}),
		generator:   minimalGoGenerator,
		goBin:       run.FindGo(),
	}
}

func minimalGoGenerator(input gen.Input) (result gen.Result, err error) {
	err = run.Go(input.WorkingDir, "mod", "init", "usql")
	if err != nil {
		return
	}
	err = os.WriteFile(filepath.Join(input.WorkingDir, "main.go"), []byte(main), 0644)
	return
}
