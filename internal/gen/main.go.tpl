package main

import (
	"github.com/samber/lo"
	"usql/gen"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/shell"
)

{{range $val := .Imports}}
import _ "{{$val}}"
{{end}}

func main() {
{{if .Imports}}
	newDrivers := gen.RegisterNewDrivers(lo.Keys(drivers.Available()))
	for _, driver := range newDrivers {
		drivers.Register(driver, drivers.Driver{})
	}
{{end}}
	shell.Run()
}
