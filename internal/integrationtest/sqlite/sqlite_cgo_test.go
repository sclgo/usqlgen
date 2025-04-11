//go:build cgo

package sqlite

import (
	"testing"

	"github.com/sclgo/usqlgen/internal/gen"
	it "github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestSqliteCgo(t *testing.T) {
	inp := gen.Input{}

	tmpDir := t.TempDir()
	inp.WorkingDir = tmpDir

	err := inp.All()
	require.NoError(t, err)

	output := it.RunGeneratedUsql(t, "sqlite3::memory:", `select sqlite_version()`, tmpDir, "no_moderncsqlite")
	require.Contains(t, output, "sqlite_version()")
}
