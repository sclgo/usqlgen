package cosmos

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/murfffi/gorich/fi"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestCosmos(t *testing.T) {
	fi.SkipLongTest(t)

	inp := gen.Input{
		Imports: []string{"github.com/microsoft/gocosmos"},
	}

	tmpDir := t.TempDir()
	inp.WorkingDir = tmpDir

	err := inp.All()
	require.NoError(t, err)

	cmd := exec.Command("go", "run", "-mod=mod", "-tags", integrationtest.NoBaseTag, ".", "gocosmos:AccountEndpoint=https://localhost;AccountKey=test", "-c", `LIST DATABASES;`)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = io.MultiWriter(&buf, os.Stderr)
	cmd.Env = slices.DeleteFunc(os.Environ(), func(s string) bool {
		return strings.HasPrefix(s, "GO")
	})
	err = cmd.Run()
	output := buf.String()
	require.ErrorContains(t, err, "exit status 1")
	require.Contains(t, output, `Get "https://localhost/dbs"`)
}
