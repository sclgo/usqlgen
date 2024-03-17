package gen_test

import (
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegisterNewDrivers(t *testing.T) {
	var drivers = []string{"hello"}
	newDrivers := gen.RegisterNewDrivers(drivers)
	require.Equal(t, drivers, newDrivers)
}
