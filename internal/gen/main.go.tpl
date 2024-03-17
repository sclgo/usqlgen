package main

import (
	"github.com/samber/lo"
	"github.com/sclgo/usqlgen/pkg/dbmgr"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/shell"
)

{{range $val := .Imports}}
import _ "{{$val}}"
{{end}}

func main() {
{{if .Imports}}
	newDrivers := dbmgr.RegisterNewDrivers(lo.Keys(drivers.Available()))
	for _, driver := range newDrivers {
		drivers.Register(driver, drivers.Driver{})
	}
{{end}}
	shell.Run()
}
