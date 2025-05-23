package databend_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/murfffi/gorich/fi"
	"github.com/samber/lo"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const Username = "databend"
const Password = "databend"

const query = `SELECT avg(number) as average FROM numbers(100000000)`

func GetDsn(ctx context.Context, c testcontainers.Container) string {
	port := lo.Must(c.MappedPort(ctx, "8000/tcp"))
	return fmt.Sprintf("http://%s:%s@%s:%d/default?sslmode=disable&presigned_url_disabled=true", Username, Password, lo.Must(c.Host(ctx)), port.Int())
}

func Setup(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		// Right now databend-go transaction prepared statements work on nightly but not on latest
		Image:        "datafuselabs/databend:nightly",
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
	c := fi.NoError(Setup(ctx)).Require(t)

	t.Cleanup(func() {
		assert.NoError(t, c.Terminate(ctx))
	})
	dsn := GetDsn(ctx, c)

	inp := gen.Input{
		Imports: []string{"github.com/datafuselabs/databend-go"},
	}

	t.Run("basic query", func(t *testing.T) {
		integrationtest.CheckGenAll(t, inp, "databend:"+dsn, query)
	})

	t.Run("copy", func(t *testing.T) {
		tmpDir := t.TempDir()
		inp.WorkingDir = tmpDir

		err := inp.All()
		require.NoError(t, err)

		output := integrationtest.RunGeneratedUsql(t, "databend:"+dsn, "create table dest(col1 string, col2 string);", tmpDir)
		require.Contains(t, output, "CREATE TABLE")

		destExpression := "INSERT INTO dest VALUES (?, ?)"
		copyCmd := fmt.Sprintf(`\copy csvq:. databend:%s 'select string(1), string(2)' '%s'`, dsn, destExpression)
		output = integrationtest.RunGeneratedUsql(t, "", copyCmd, tmpDir, "csvq")
		require.Contains(t, output, "COPY")
		output = integrationtest.RunGeneratedUsql(t, "databend:"+dsn, "select * from dest", tmpDir)
		require.Contains(t, output, "(1 row)")
	})
}
