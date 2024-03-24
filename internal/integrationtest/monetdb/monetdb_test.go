package monetdb

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/sclgo/usqlgen/pkg/fi"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"

	_ "github.com/MonetDB/MonetDB-Go/src"
)

const Username = "monetdb"
const Password = "monetdb"
const dbPort = "50000/tcp"

func TestMonetdb(t *testing.T) {
	integrationtest.IntegrationOnly(t)
	ctx := context.Background()
	c := fi.NoError(Setup(ctx)).Require(t)

	defer integrationtest.Terminate(ctx, t, c)
	dsn := GetDsn(ctx, c)
	integrationtest.SanityPing(ctx, t, dsn, "monetdb")

	inp := gen.Input{
		Imports: []string{"github.com/MonetDB/MonetDB-Go/v2"},
	}

	integrationtest.CheckGenAll(t, inp, "monetdb:"+dsn, "select 1")
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
