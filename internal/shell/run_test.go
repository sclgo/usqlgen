package shell_test

import (
	"bytes"
	"github.com/sclgo/usqlgen/internal/shell"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerate(t *testing.T) {
	var buf bytes.Buffer
	shell.RunArgs([]string{"usqlgen", "generate", "--help"}, &buf, nil)
	require.Contains(t, buf.String(), "generate")
	require.Contains(t, buf.String(), "import")
}
