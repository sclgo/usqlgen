package monetdb

import (
	"context"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/integrationtest"
	"github.com/sclgo/usqlgen/pkg/fi"
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

	defer integrationtest.Terminate(ctx, t, c)
	dsn := GetDsn(ctx, t, c)

	t.Run("kprotoss driver", func(t *testing.T) {
		inp := gen.Input{
			Imports: []string{"github.com/kprotoss/go-impala"},
		}
		integrationtest.CheckGenAll(t, inp, "impala:"+dsn, "select 'Hello World'")
	})

	t.Run("kenshaw driver", func(t *testing.T) {
		inp := gen.Input{
			Replaces: []string{"github.com/bippio/go-impala=github.com/kenshaw/go-impala@latest"},
		}
		integrationtest.CheckGenAll(t, inp, dsn, "select 'Hello World'", "impala")
	})

}

func GetDsn(ctx context.Context, t *testing.T, c testcontainers.Container) string {
	port := fi.NoError(c.MappedPort(ctx, dbPort)).Require(t).Port()
	host := fi.NoError(c.Host(ctx)).Require(t)
	u := &url.URL{
		Scheme: "impala",
		Host:   net.JoinHostPort(host, port),
	}
	t.Log("url", u.String())
	return u.String()
}

func Setup(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "apache/kudu:impala-latest",
		ExposedPorts: []string{dbPort},
		Cmd:          []string{"impala"},
		WaitingFor:   wait.ForLog("Starting statestore subscriber service"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
