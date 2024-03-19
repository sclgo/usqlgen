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

	inp := gen.Input{
		Imports:  []string{"github.com/MonetDB/MonetDB-Go/v2"},
		Replaces: []string{"github.com/MonetDB/MonetDB-Go=github.com/sclgo/MonetDB-Go@latest"},
	}
	tmpDir, err := os.MkdirTemp("/tmp", "usqltest")
	require.NoError(t, err)
	defer fi.MustF(fi.Bind(os.RemoveAll, tmpDir), t)
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

	require.Contains(t, buf.String(), "monetdb")
}
