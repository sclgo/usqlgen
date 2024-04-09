package monetdb

import (
	"context"
	"fmt"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net"
	"net/url"
	"testing"
)

const dbPort = "21050/tcp"

func TestImpala(t *testing.T) {
	integrationtest.IntegrationOnly(t)
	ctx := context.Background()
	c := fi.NoError(Setup(ctx)).Require(t)
	defer fi.NoErrorF(fi.Bind(c.Terminate, ctx), t)

	dsn := GetDsn(ctx, t, c)

	t.Run("kprotoss driver", func(t *testing.T) {
		inp := gen.Input{
			Imports: []string{"github.com/kprotoss/go-impala"},
		}
		t.Run("select", func(t *testing.T) {
			integrationtest.CheckGenAll(t, inp, "impala:"+dsn, "select 'Hello World'")
		})

	})

	t.Run("sclgo driver", func(t *testing.T) {
		inp := gen.Input{
			Imports:  []string{"github.com/bippio/go-impala"},
			Replaces: []string{"github.com/bippio/go-impala=github.com/sclgo/go-impala@master"},
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

			//impala.DefaultOptions.LogOut = os.Stdout
			//db, err := sql.Open("impala", dsn)
			//require.NoError(t, err)
			//defer sclerr.CloseQuietly(db)
			//_, err = db.Exec(tableDdl)
			//require.NoError(t, err)
			//_, err = db.Exec("insert into default.dest values ('1', '2')")
			//require.NoError(t, err)

			output := integrationtest.RunGeneratedUsql(t, "impala:"+dsn, tableDdl, tmpDir)
			require.Contains(t, output, "CREATE TABLE")

			destExpression := "INSERT INTO dest VALUES (?, ?)"
			copyCmd := fmt.Sprintf(`\copy csvq:. impala:%s 'select string(1), string(2)' '%s'`, dsn, destExpression)
			output = integrationtest.RunGeneratedUsql(t, "", copyCmd, tmpDir)
			require.Contains(t, output, "COPY")

			output = integrationtest.RunGeneratedUsql(t, "impala:"+dsn, "select * from dest", tmpDir)

			//INSERT does not work on the combination current version of the driver and
			//the docker environment. Writing to tables backed by local files requires passing the user
			//which doesn't seem to happen when useLdap = false.
			//require.Contains(t, output, "(1 row)")
			require.Contains(t, output, "(0 rows)")
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
		WaitingFor:   wait.ForLog("Impala has started."),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
