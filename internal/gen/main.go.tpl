package main

{{if .Imports}}
import (
	"github.com/samber/lo"
	"github.com/xo/usql/gen"
	"github.com/xo/usql/drivers"
)
{{end}}

{{range $val := .Imports}}
import _ "{{$val}}"
{{end}}

func main() {
{{if .Imports}}
	newDrivers := gen.RegisterNewDrivers(lo.Keys(drivers.Available()))
	for _, driver := range newDrivers {
		drivers.Register(driver, drivers.Driver{
		    Copy: gen.BuildSimpleCopy(gen.FixedPlaceholder("?")),
		})
	}
{{end}}
	origMain()
}
