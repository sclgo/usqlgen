package databend_test

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
	c := fi.NoError(Setup(ctx)).Require(t)

	defer integrationtest.Terminate(ctx, t, c)
	dsn := GetDsn(ctx, c)

	integrationtest.SanityPing(ctx, t, dsn, "databend")

	inp := gen.Input{
		Imports: []string{"github.com/datafuselabs/databend-go"},
	}

	integrationtest.CheckGenAll(t, inp, "databend:"+dsn, query)
}
