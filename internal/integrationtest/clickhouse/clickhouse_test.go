//go:build linux

package clickhouse_test

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/murfffi/gorich/fi"
	"github.com/murfffi/gorich/sclerr"
	"github.com/sclgo/usqlgen/internal/gen"
	it "github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestClickhouse(t *testing.T) {
	fi.SkipLongTest(t)
	stderrPipe, stop := startClickhouse(t)
	defer stop() // don't use t.Cleanup here so stop gets executed even on timeout-induced panic

	scanner := bufio.NewScanner(stderrPipe)
	timer := time.NewTimer(20 * time.Second)

	ch := make(chan bool)
	go func() {
		ch <- waitForClickhouse(scanner)
	}()

	select {
	case <-timer.C:
		require.Fail(t, "clickhouse server didn't start in time")
	case ready := <-ch:
		require.True(t, ready, "clickhouse server failed to start")
	}
	timer.Stop()

	inp := gen.Input{
		Imports: []string{"github.com/mailru/go-clickhouse/v2"},
	}

	dsn := "chhttp:http://localhost:8123/"

	t.Run("basic query", func(t *testing.T) {
		it.CheckGenAll(t, inp, dsn, `select 'Hello World'`)
	})

	t.Run("copy", func(t *testing.T) {
		tmpDir := t.TempDir()
		inp.WorkingDir = tmpDir

		err := inp.All()
		require.NoError(t, err)

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

func waitForClickhouse(scanner *bufio.Scanner) bool {
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Application: Ready for connections") {
			return true
		}
	}
	return false
}

func startClickhouse(t *testing.T) (io.ReadCloser, func()) {
	dir := t.TempDir()
	clickhouseBinName := filepath.Join(dir, "clickhouse")
	clickhouseBin, err := os.OpenFile(clickhouseBinName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	require.NoError(t, err)
	defer sclerr.CloseQuietly(clickhouseBin)

	resp, err := http.Get("https://builds.clickhouse.com/master/amd64/clickhouse")
	require.NoError(t, err)
	defer sclerr.CloseQuietly(resp.Body)

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
