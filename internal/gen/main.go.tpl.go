package gen

// mainTpl contains the template as a constant as opposed to using go:embed with a separate
// file because this way staticcheck (and by extension golangci-lint) can pick up errors
// with SA1001
const mainTpl = `
package main

import (
	"maps"
	"slices"
	"github.com/xo/usql/gen"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/env"
	"github.com/xo/dburl"

	_ "github.com/xo/usql/internal"
)

{{range $val := .Imports}}
import _ "{{$val}}"
{{end}}

func main() {
	newDrivers := gen.RegisterNewDrivers(slices.Collect(maps.Keys(drivers.Available())))
	for _, driver := range newDrivers {
		drivers.Register(driver, drivers.Driver{
			Copy: gen.BuildSimpleCopy(gen.FixedPlaceholder("?")),
			{{if not .IncludeSemicolon}}
			Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
				sqlstr = gen.SemicolonEndRE.ReplaceAllString(sqlstr, "")
				typ, q := drivers.QueryExecType(prefix, sqlstr)
				return typ, sqlstr, q, nil
			},
			{{end}}
		})
	}
	// The default prompt is sometimes too long for DBs with opaque URLs
	env.Set("PROMPT1", "%S%N%m%R%# ")
	origMain()
}
`
