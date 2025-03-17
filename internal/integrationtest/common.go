package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/pkg/sclerr"
	"github.com/stretchr/testify/require"
)

const NoBaseTag = "no_base"

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
	// speed up build, add some tag like "base" to disable
	if tags == nil {
		tags = []string{NoBaseTag}
	}

	cmd := exec.Command("go", "run", "-mod=mod", "-tags", strings.Join(tags, ","), ".", dsn, "-c", command)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	cmd.Env = slices.DeleteFunc(os.Environ(), func(s string) bool {
		return strings.HasPrefix(s, "GO")
	})
	err := cmd.Run()
	output := buf.String()
	return output, err
}
