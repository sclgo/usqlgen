package gen_test

import (
	"bytes"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"testing"
)

func TestInput_Main(t *testing.T) {
	inp := gen.Input{
		Imports: []string{"hello/hello"},
	}
	buf := bytes.Buffer{}
	require.NoError(t, inp.Main(&buf))
	require.Contains(t, buf.String(), "hello/hello")
}

func TestInput_All(t *testing.T) {
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
	cmd.Stdout = &buf
	err = cmd.Run()
	require.NoError(t, err)

	require.Contains(t, buf.String(), "databend")
}
