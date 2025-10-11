package integrationtest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/murfffi/gorich/fi"
	"github.com/murfffi/gorich/lang"
	"github.com/sclgo/usqlgen/internal/shell"
	"github.com/stretchr/testify/require"
)

// Tests of the overall build process

// TestVersion is an integration test for --version in the produced binary
// This test complements the unit test in shell/cmd_build_test.go. The unit test
// validates only if the version is injected in the build process but only this
// test confirms that usql will use it.
func TestVersion(t *testing.T) {
	IntegrationOnly(t)

	tmpDir := t.TempDir()
	currentWorkDir := fi.NoError(os.Getwd()).Require(t)
	defer fi.NoErrorF(lang.Bind(os.Chdir, currentWorkDir), t)
	require.NoError(t, os.Chdir(tmpDir))

	cmds := shell.NewCommands(nil)
	cmds.BuildCmd.USQLModule = "github.com/xo/usql"
	testVersion := "v0.19.14"
	cmds.BuildCmd.USQLVersion = testVersion
	cmds.BuildCmd.Globals.PassthroughArgs = []string{"-tags", "no_base"}

	err := cmds.BuildCmd.Action(nil)
	require.NoError(t, err)

	cmd := exec.Command("./usql", "--version")
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())

	require.Equal(t, fmt.Sprintf("usql %s_usqlgen", testVersion), strings.TrimSpace(buf.String()))
}
