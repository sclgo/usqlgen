package databend_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/samber/lo"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"os/exec"
	"testing"
)

import _ "github.com/datafuselabs/databend-go"

const Username = "databend"
const Password = "databend"

const query = `SELECT avg(number) as average FROM numbers(100000000)`

func GetDsn(ctx context.Context, c testcontainers.Container) string {
	port := lo.Must(c.MappedPort(ctx, "8000/tcp"))
	return fmt.Sprintf("http://%s:%s@%s:%d/default?sslmode=disable", Username, Password, lo.Must(c.Host(ctx)), port.Int())
}

func Setup(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "datafuselabs/databend",
		ExposedPorts: []string{"8000/tcp"},
		Env: map[string]string{
			"QUERY_DEFAULT_USER":     Username,
			"QUERY_DEFAULT_PASSWORD": Password,
		},
		WaitingFor: wait.ForLog("Databend Metasrv started"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func TestDatabend(t *testing.T) {
	integrationtest.IntegrationOnly(t)
	ctx := context.Background()
	c := fi.Must(Setup(ctx))(t)

	defer integrationtest.Terminate(ctx, t, c)
	dsn := GetDsn(ctx, c)

	sanityPing(t, dsn, ctx)

	tmpDir, err := os.MkdirTemp("/tmp", "usqltest")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	inp := gen.Input{
		Imports:    []string{"github.com/datafuselabs/databend-go"},
		WorkingDir: tmpDir,
	}

	err = inp.All()
	require.NoError(t, err)

	cmd := exec.Command("go", "run", ".", "databend:"+dsn, "-c", query)
	cmd.Dir = tmpDir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err = cmd.Run()
	require.NoError(t, err)

	t.Log(buf.String())
	require.Contains(t, buf.String(), "(1 row)")
}

func sanityPing(t *testing.T, dsn string, ctx context.Context) {
	db, err := sql.Open("databend", dsn)
	require.NoError(t, err)
	defer db.Close()
	err = db.PingContext(ctx)
	require.NoError(t, err)
}
