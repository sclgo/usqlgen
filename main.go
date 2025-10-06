package main

import (
	// for USR1, USR2 handling e.g. "pkill -USR1 usqlgen" prints stacktrace
	// see docs for details
	_ "github.com/tam7t/sigprof"

	"github.com/sclgo/usqlgen/internal/shell"
)

func main() {
	shell.Run()
}
