package gen_test

import (
	"database/sql"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegisterNewDrivers(t *testing.T) {
	newDrivers := gen.RegisterNewDrivers(sql.Drivers())
	require.Empty(t, newDrivers)
}
