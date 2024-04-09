package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/shell"
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
	tmpDir := t.TempDir()
	inp.WorkingDir = tmpDir

	err := inp.All()
	require.NoError(t, err)

	output := RunGeneratedUsql(t, dsn, command, tmpDir, tags...)
	require.Contains(t, output, "(1 row)")
}

func RunGeneratedUsql(t *testing.T, dsn string, command string, tmpDir string, tags ...string) string {
	t.Logf("Running cmd %s with dsn %s", command, dsn)
	output, err := RunGeneratedUsqlE(dsn, command, tmpDir, tags...)
	require.NoError(t, err)
	return output
}

func RunGeneratedUsqlE(dsn string, command string, tmpDir string, tags ...string) (string, error) {
	cmd := exec.Command("go", "run", "-tags", strings.Join(tags, ","), ".", dsn, "-c", command)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	output := buf.String()
	return output, err
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
