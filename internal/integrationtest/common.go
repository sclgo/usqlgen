package integrationtest

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"os"
	"strings"
	"testing"
)

func Terminate(ctx context.Context, t *testing.T, c testcontainers.Container) {
	require.NoError(t, c.Terminate(ctx))
}

func IntegrationOnly(t *testing.T) {
	if strings.ToLower(os.Getenv("SUITE")) != "integration" {
		t.SkipNow()
	}
}
