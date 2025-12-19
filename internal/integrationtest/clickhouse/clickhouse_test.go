//go:build linux

package clickhouse_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/murfffi/gorich/fi"
	"github.com/murfffi/gorich/helperr"
	"github.com/sclgo/usqlgen/internal/gen"
	it "github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestClickhouse(t *testing.T) {
	fi.SkipLongTest(t)
	stderrPipe, stop := startClickhouse(t)
	defer stop() // don't use t.Cleanup here so stop gets executed even on timeout-induced panic

	fi.RequireReaderContains(t, stderrPipe, "Application: Ready for connections", 20*time.Second, "clickhouse log")
	_ = stderrPipe.Close()

	inp := gen.Input{
		Imports: []string{"github.com/mailru/go-clickhouse/v2"},
	}

	dsn := "chhttp:http://localhost:8123/"

	tmpDir := t.TempDir()
	inp.WorkingDir = tmpDir

	err := inp.All()
	require.NoError(t, err)

	t.Run("basic query", func(t *testing.T) {
		it.RunGeneratedUsql(t, dsn, `select 'Hello World'`, tmpDir)
	})

	t.Run("metadata", func(t *testing.T) {
		_, _ = it.RunGeneratedUsqlE(dsn, "create table tmptmp(col1 varchar, col2 varchar) PRIMARY KEY col1;", tmpDir, nil)
		output := it.RunGeneratedUsql(t, dsn, `\dtS`, tmpDir)
		require.Contains(t, output, "tmptmp")
	})

	t.Run("copy", func(t *testing.T) {
		_, _ = it.RunGeneratedUsqlE(dsn, "create table dest(col1 varchar, col2 varchar) PRIMARY KEY col1;", tmpDir, nil)
		// ignoring error because usql -c, for some reason, fails if RowsAffected returns error
		// chhttp returns error from RowsAffected because it doesn't support it

		destExpression := "INSERT INTO dest VALUES (?, ?)"
		copyCmd := fmt.Sprintf(`\copy csvq:. %s 'select string(1), string(2)' '%s'`, dsn, destExpression)
		output := it.RunGeneratedUsql(t, "", copyCmd, tmpDir, "csvq")
		require.Contains(t, output, "COPY")
		output = it.RunGeneratedUsql(t, dsn, "select * from dest", tmpDir)
		require.Contains(t, output, "(1 row)")
	})
}

func startClickhouse(t *testing.T) (io.ReadCloser, func()) {
	dir := t.TempDir()
	clickhouseBinName := filepath.Join(dir, "clickhouse")
	clickhouseBin, err := os.OpenFile(clickhouseBinName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	require.NoError(t, err)
	defer helperr.CloseQuietly(clickhouseBin)

	resp, err := http.Get("https://builds.clickhouse.com/master/amd64/clickhouse")
	require.NoError(t, err)
	defer helperr.CloseQuietly(resp.Body)

	_, err = io.Copy(clickhouseBin, resp.Body)
	require.NoError(t, err)
	require.NoError(t, clickhouseBin.Close())

	cmd := exec.Command(clickhouseBinName, "server")
	cmd.Dir = dir
	pipe := fi.NoError(cmd.StderrPipe()).Require(t)
	err = cmd.Start()
	require.NoError(t, err)

	return pipe, func() {
		t.Log("Stopping clickhouse")
		require.NoError(t, cmd.Process.Kill())
		_ = cmd.Wait()
	}
}
