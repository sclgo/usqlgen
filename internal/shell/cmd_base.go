package shell

import (
	"github.com/urfave/cli/v2"
)

type GlobalParams struct {
	Verbose         bool
	PassthroughArgs []string
}

type CommandBase struct {
	Globals *GlobalParams
}

func (c *CommandBase) MakeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "enabled detailed, debug output",
			Aliases:     []string{"v"},
			Destination: &c.Globals.Verbose,
		},
	}
}
