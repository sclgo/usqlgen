package sqlite

import (
	"runtime"
	"testing"

	"github.com/sclgo/usqlgen/internal/gen"
	it "github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

// This file can't contain any tests that require CGO.
// The tests that require CGO must be skipped - not failed - on systems without CGO.
// CGO tests go in sqlite_cgo_test.go

func TestSqlite_NoCgo(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping on No CGO test on MacOS due to https://github.com/jeandeaual/go-locale/issues/30#issuecomment-2798583087")
	}
	noCgoEnv := []string{"CGO_ENABLED=0"}

	t.Run("base", func(t *testing.T) {
		// Confirm that base driver set supports sqlite3 schema (via a replacement) even
		// if original sqlite3 is not available due to no CGO
		tag := "base"
		inp := gen.Input{}

		tmpDir := t.TempDir()
		inp.WorkingDir = tmpDir

		err := inp.All()
		require.NoError(t, err)

		_, err = it.RunGeneratedUsqlE("", `\drivers`, tmpDir, noCgoEnv, tag)
		require.NoError(t, err)

		output, err := it.RunGeneratedUsqlE("sqlite3::memory:", `select sqlite_version()`, tmpDir, noCgoEnv, tag)
		require.NoError(t, err)
		require.Contains(t, output, "sqlite_version()")
	})

	t.Run("no_replacement", func(t *testing.T) {
		// Check that replacement logic obeys the tags that block sqlite3 replacement
		for _, tag := range []string{"no_base", "no_moderncsqlite", "no_sqlite3"} {
			t.Run(tag, func(t *testing.T) {
				inp := gen.Input{}

				tmpDir := t.TempDir()
				t.Log(tmpDir)
				inp.WorkingDir = tmpDir

				err := inp.All()
				require.NoError(t, err)

				_, err = it.RunGeneratedUsqlE("", `\drivers`, tmpDir, noCgoEnv, tag)
				require.NoError(t, err)

				_, err = it.RunGeneratedUsqlE("sqlite3::memory:", `select sqlite_version()`, tmpDir, noCgoEnv, tag)
				require.Error(t, err)
			})
		}
	})

	t.Run("keepcgo", func(t *testing.T) {
		// Check that build fails if keepcgo is active but CGO is not available
		inp := gen.Input{
			KeepCgo: true,
		}

		tmpDir := t.TempDir()
		inp.WorkingDir = tmpDir

		err := inp.All()
		require.NoError(t, err)

		_, err = it.RunGeneratedUsqlE("", `\drivers`, tmpDir, noCgoEnv, "base")
		require.Error(t, err)
	})
}
