package gen_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/samber/lo"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/stretchr/testify/require"
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
		require.Contains(t, stdoutStr, "monetdb [mo]")
		require.NotContains(t, stdoutStr, "github.com/MonetDB/MonetDB-Go/v2 v2.0.1")
		// incorrectly re-registered sqlserver, which used to be a side-effect of imports
		require.NotContains(t, stdoutStr, "mssql [ms]")
	})

	t.Run("replaces", func(t *testing.T) {
		inp := gen.Input{
			Imports:  []string{"github.com/MonetDB/MonetDB-Go/v2"},
			Replaces: []string{"github.com/MonetDB/MonetDB-Go/v2=github.com/sclgo/MonetDB-Go/v2@fbbd00a"},
		}
		stdoutStr := runGenAll(t, inp)

		require.Contains(t, stdoutStr, "monetdbscl")
	})

	t.Run("gets", func(t *testing.T) {
		inp := gen.Input{
			Imports: []string{"github.com/MonetDB/MonetDB-Go/v2"},
			Gets:    []string{"github.com/MonetDB/MonetDB-Go/v2@v2.0.1"},
		}
		stdoutStr := runGenAll(t, inp)

		require.Contains(t, stdoutStr, "github.com/MonetDB/MonetDB-Go/v2 v2.0.1")
	})
}

func runGenAll(t *testing.T, inp gen.Input) string {
	tmpDir := t.TempDir()
	defer fi.NoErrorF(fi.Bind(os.RemoveAll, tmpDir), t)
	inp.WorkingDir = tmpDir

	err := inp.All()
	require.NoError(t, err)

	cmd := exec.Command("go", "run", ".", "-c", `\drivers`)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err, buf.String())

	cmd = exec.Command("go", "list", "-m", "all")
	cmd.Dir = tmpDir
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	stdoutStr := buf.String()
	require.NoError(t, err, stdoutStr)
	return stdoutStr
}

func TestInput_AllDownload(t *testing.T) {
	fi.SkipLongTest(t)
	inp := gen.Input{
		Imports:     []string{"github.com/MonetDB/MonetDB-Go/v2"},
		USQLVersion: "v0.19.14",
	}
	var err error
	tmpDir := t.TempDir()
	inp.WorkingDir = tmpDir
	result, err := inp.AllDownload()
	require.NoError(t, err)
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.True(t, lo.ContainsBy(entries, func(item os.DirEntry) bool {
		return item.Name() == "main.go"
	}))

	require.Equal(t, inp.USQLVersion, result.DownloadedUsqlVersion)
}
