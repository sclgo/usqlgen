package gen_test

import (
	"bytes"
	"github.com/sclgo/usqlgen/internal/gen"
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
	if testing.Short() {
		t.SkipNow()
	}
	inp := gen.Input{
		Imports: []string{"github.com/datafuselabs/databend-go"},
	}
	tmpDir, err := os.MkdirTemp("/tmp", "usqltest")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	inp.WorkingDir = tmpDir

	err = inp.All()
	require.NoError(t, err)

	cmd := exec.Command("go", "run", ".", "-c", `\drivers`)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err, buf.String())

	require.Contains(t, buf.String(), "databend")
}
