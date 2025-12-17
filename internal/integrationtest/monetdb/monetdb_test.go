package monetdb

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

const Username = "monetdb"
const Password = "monetdb"
const dbPort = "50000/tcp"

func TestMonetdb(t *testing.T) {
	integrationtest.IntegrationOnly(t)
	ctx := context.Background()
	c := fi.NoError(Setup(ctx)).Require(t)
	t.Cleanup(func() {
		assert.NoError(t, c.Terminate(ctx))
	})

	dsn := GetDsn(ctx, c)

	inp := gen.Input{
		Imports: []string{"github.com/MonetDB/MonetDB-Go/v2"},
	}

	tmpDir := t.TempDir()
	inp.WorkingDir = tmpDir

	err := inp.All()
	require.NoError(t, err)

	t.Run("basic", func(t *testing.T) {
		integrationtest.RunGeneratedUsql(t, "monetdb:"+dsn, "select 1", tmpDir)
	})
	t.Run("alias", func(t *testing.T) {
		integrationtest.RunGeneratedUsql(t, "mo:"+dsn, "select 1", tmpDir)
	})
	t.Run("list tables", func(t *testing.T) {
		// MonetDB supports information schema, but usql's InformationSchema implementation is not
		// compatible with the driver. usql InformationSchema expects that the driver converts NULL values
		// to zero values when scanning into a non-reference type.
		integrationtest.RunGeneratedUsql(t, "monetdb:"+dsn, `\dtS`, tmpDir)
	})

}

func GetDsn(ctx context.Context, c testcontainers.Container) string {
	port := lo.Must(c.MappedPort(ctx, dbPort))
	return fmt.Sprintf("%s:%s@%s:%d/monetdb", Username, Password, lo.Must(c.Host(ctx)), port.Int())
}

func Setup(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "monetdb/monetdb",
		ExposedPorts: []string{dbPort},
		Env: map[string]string{
			"MDB_DB_ADMIN_PASS": Password,
		},
		WaitingFor: wait.ForLog("Starting MonetDB daemon"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
