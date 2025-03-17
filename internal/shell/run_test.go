package shell_test

import (
	"bytes"
	"testing"

	"github.com/sclgo/usqlgen/internal/shell"
	"github.com/stretchr/testify/require"
)

func TestRunArgs(t *testing.T) {
	var buf bytes.Buffer
	shell.RunArgs([]string{"usqlgen", "generate", "--help"}, &buf, nil)
	require.Contains(t, buf.String(), "generate")
	require.Contains(t, buf.String(), "import")
}
