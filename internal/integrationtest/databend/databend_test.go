package databend_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/samber/lo"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/sclgo/usqlgen/pkg/sclerr"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path/filepath"
	"testing"
)

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
	c := fi.NoError(Setup(ctx)).Require(t)

	defer integrationtest.Terminate(ctx, t, c)
	dsn := GetDsn(ctx, c)

	inp := gen.Input{
		Imports: []string{"github.com/datafuselabs/databend-go"},
	}

	t.Run("basic query", func(t *testing.T) {
		integrationtest.CheckGenAll(t, inp, "databend:"+dsn, query)
	})

	t.Run("copy", func(t *testing.T) {
		t.Skip(`the following doesn't work because usql CopyWithInsert starts a transaction,
but databend doesn't support multi-statement transactions. The databend driver (which appears
to be a fork of clickhouse-go) issues a BEGIN statement on db.BeginTx even though the DB
doesn't support it.`)

		tmpDir := integrationtest.MakeTempDir(t)
		defer fi.NoErrorF(fi.Bind(os.RemoveAll, tmpDir), t)
		inp.WorkingDir = tmpDir

		err := inp.All()
		require.NoError(t, err)

		csvFile, err := os.Create(filepath.Join(tmpDir, "data.csv"))
		require.NoError(t, err)
		defer sclerr.CloseQuietly(csvFile)

		csvWriter := csv.NewWriter(csvFile)
		require.NoError(t, csvWriter.Write([]string{"hello", "world"}))
		require.NoError(t, csvFile.Close())

		output := integrationtest.RunGeneratedUsql(t, "databend:"+dsn, "create table dest(col1 string, col2 string);", tmpDir)

		destExpression := "INSERT INTO dest VALUES (?, ?)"
		copyCmd := fmt.Sprintf(`\copy csvq:. databend:%s 'select * from data' '%s'`, dsn, destExpression)
		output = integrationtest.RunGeneratedUsql(t, "", copyCmd, tmpDir)
		require.Contains(t, output, "(1 row)")
	})
}
