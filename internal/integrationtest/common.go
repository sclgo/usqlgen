package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/shell"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/sclgo/usqlgen/pkg/sclerr"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func IntegrationOnly(t *testing.T) {
	if strings.ToLower(os.Getenv("SUITE")) != "integration" {
		t.Skip("This test requires env var SUITE=integration")
		t.SkipNow()
	}
}

func SanityPing(ctx context.Context, t *testing.T, dsn string, driver string) {
	db, err := sql.Open(driver, dsn)
	require.NoError(t, err)
	defer sclerr.CloseQuietly(db)
	err = db.PingContext(ctx)
	require.NoError(t, err)
}

func CheckGenAll(t *testing.T, inp gen.Input, dsn string, command string, tags ...string) {
	tmpDir := MakeTempDir(t)
	defer fi.NoErrorF(fi.Bind(os.RemoveAll, tmpDir), t)
	inp.WorkingDir = tmpDir

	err := inp.All()
	require.NoError(t, err)

	output := RunGeneratedUsql(t, dsn, command, tmpDir, tags...)
	require.Contains(t, output, "(1 row)")
}

func MakeTempDir(t *testing.T) string {
	return fi.NoError(os.MkdirTemp("/tmp", "usqltest")).Require(t)
}

func RunGeneratedUsql(t *testing.T, dsn string, command string, tmpDir string, tags ...string) string {
	cmd := exec.Command("go", "run", "-tags", strings.Join(tags, ","), ".", dsn, "-c", command)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	require.NoError(t, err)

	output := buf.String()
	t.Log(output)
	return output
}

func CheckBuildRun(t *testing.T, inp gen.Input, dsn string, query string, tags ...string) {
	cmd := &shell.BuildCommand{
		CompileCommand: shell.CompileCommand{
			CommandBase: shell.CommandBase{},
			Imports:     *cli.NewStringSlice(inp.Imports...),
			Replaces:    *cli.NewStringSlice(inp.Replaces...),
		},
	}
	cmd.Action(nil)
	// TODO complete
}
