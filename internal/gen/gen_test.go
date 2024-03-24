package gen_test

import (
	"bytes"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"os/exec"
	"testing"
)

func TestInput_Main(t *testing.T) {
	t.Run("with imports", func(t *testing.T) {
		inp := gen.Input{
			Imports: []string{"hello/hello"},
		}
		buf := bytes.Buffer{}
		require.NoError(t, inp.Main(&buf))
		require.Contains(t, buf.String(), "hello/hello")
	})

	t.Run("without imports", func(t *testing.T) {
		var inp gen.Input
		buf := bytes.Buffer{}
		require.NoError(t, inp.Main(&buf))
		require.NotContains(t, buf.String(), "RegisterNewDrivers")
	})
}

func TestInput_All(t *testing.T) {
	fi.SkipLongTest(t)
	t.Run("imports", func(t *testing.T) {
		inp := gen.Input{
			Imports: []string{"github.com/MonetDB/MonetDB-Go/v2"},
		}
		stdoutStr := runGenAll(t, inp)
		require.Contains(t, stdoutStr, "monetdb")
	})

	t.Run("replaces", func(t *testing.T) {
		inp := gen.Input{
			Imports:  []string{"github.com/MonetDB/MonetDB-Go/v2"},
			Replaces: []string{"github.com/MonetDB/MonetDB-Go/v2=github.com/sclgo/MonetDB-Go/v2@fbbd00a"},
		}
		stdoutStr := runGenAll(t, inp)

		require.Contains(t, stdoutStr, "monetdbscl")
	})
}

func runGenAll(t *testing.T, inp gen.Input) string {
	tmpDir, err := os.MkdirTemp("/tmp", "usqltest")
	require.NoError(t, err)
	defer fi.NoErrorF(fi.Bind(os.RemoveAll, tmpDir), t)
	inp.WorkingDir = tmpDir

	err = inp.All()
	require.NoError(t, err)

	cmd := exec.Command("go", "run", ".", "-c", `\drivers`)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	stdoutStr := buf.String()
	require.NoError(t, err, stdoutStr)
	return stdoutStr
}
