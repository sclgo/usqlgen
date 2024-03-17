package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func Terminate(ctx context.Context, t *testing.T, c testcontainers.Container) {
	require.NoError(t, c.Terminate(ctx))
}

func IntegrationOnly(t *testing.T) {
	if strings.ToLower(os.Getenv("SUITE")) != "integration" {
		t.SkipNow()
	}
}

func SanityPing(ctx context.Context, t *testing.T, dsn string, driver string) {
	db, err := sql.Open(driver, dsn)
	require.NoError(t, err)
	defer db.Close()
	err = db.PingContext(ctx)
	require.NoError(t, err)
}

func CheckGenAll(t *testing.T, inp gen.Input, driver string, dsn string, query string) {
	tmpDir, err := os.MkdirTemp("/tmp", "usqltest")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	inp.WorkingDir = tmpDir

	err = inp.All()
	require.NoError(t, err)

	cmd := exec.Command("go", "run", ".", driver+":"+dsn, "-c", query)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err)

	t.Log(buf.String())
	require.Contains(t, buf.String(), "(1 row)")
}
