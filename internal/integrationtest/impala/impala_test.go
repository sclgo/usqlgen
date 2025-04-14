package monetdb

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/murfffi/gorich/fi"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const dbPort = "21050/tcp"

func TestImpala(t *testing.T) {
	integrationtest.IntegrationOnly(t)
	ctx := context.Background()
	c := fi.NoError(Setup(ctx)).Require(t)
	t.Cleanup(func() {
		assert.NoError(t, c.Terminate(ctx))
	})

	dsn := GetDsn(ctx, t, c)

	t.Run("sclgo driver", func(t *testing.T) {
		inp := gen.Input{
			Imports: []string{"github.com/sclgo/impala-go"},
		}

		t.Run("select", func(t *testing.T) {
			integrationtest.CheckGenAll(t, inp, "impala:"+dsn, "select 'Hello World'")
		})

		t.Run("copy", func(t *testing.T) {
			tmpDir := t.TempDir()
			inp.WorkingDir = tmpDir

			err := inp.All()
			require.NoError(t, err)

			tableDdl := "create table default.dest(col1 string, col2 string) STORED AS PARQUET;"

			output := integrationtest.RunGeneratedUsql(t, "impala:"+dsn, tableDdl, tmpDir)
			require.Contains(t, output, "CREATE TABLE")

			destExpression := "INSERT INTO dest VALUES (?, ?)"
			copyCmd := fmt.Sprintf(`\copy csvq:. impala:%s 'select string(1), string(2)' '%s'`, dsn, destExpression)
			output = integrationtest.RunGeneratedUsql(t, "", copyCmd, tmpDir, "csvq")
			require.Contains(t, output, "COPY")

			output = integrationtest.RunGeneratedUsql(t, "impala:"+dsn, "select * from dest", tmpDir)

			require.Contains(t, output, "(1 row)")
		})

	})

	t.Run("kenshaw driver", func(t *testing.T) {
		inp := gen.Input{
			Replaces: []string{"github.com/bippio/go-impala=github.com/kenshaw/go-impala@master"},
		}

		t.Run("select", func(t *testing.T) {
			integrationtest.CheckGenAll(t, inp, dsn, "select 'Hello World'", "impala")
		})
	})

}

func GetDsn(ctx context.Context, t *testing.T, c testcontainers.Container) string {
	port := fi.NoError(c.MappedPort(ctx, dbPort)).Require(t).Port()
	host := fi.NoError(c.Host(ctx)).Require(t)
	u := &url.URL{
		Scheme: "impala",
		Host:   net.JoinHostPort(host, port),
		User:   url.User("impala"),
	}
	t.Log("url", u.String())
	return u.String()
}

func Setup(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "apache/kudu:impala-latest",
		ExposedPorts: []string{dbPort},
		Cmd:          []string{"impala"},
		WaitingFor:   wait.ForLog("Impala has started.").WithStartupTimeout(3 * time.Minute),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
