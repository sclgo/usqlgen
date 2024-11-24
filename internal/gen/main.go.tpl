package main

{{if .Imports}}
import (
	"maps"
	"slices"
	"github.com/xo/usql/gen"
	"github.com/xo/usql/drivers"
)
{{end}}

{{range $val := .Imports}}
import _ "{{$val}}"
{{end}}

func main() {
{{if .Imports}}
	newDrivers := gen.RegisterNewDrivers(slices.Collect(maps.Keys(drivers.Available())))
	for _, driver := range newDrivers {
		drivers.Register(driver, drivers.Driver{
		    Copy: gen.BuildSimpleCopy(gen.FixedPlaceholder("?")),
		})
	}
{{end}}
	origMain()
}
