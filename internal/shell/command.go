package shell

import (
	"github.com/ansel1/merry/v2"
	"github.com/urfave/cli/v2"
	"os"
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
			Destination: &c.Globals.Verbose,
		},
	}
}

func (*CommandBase) Action(*cli.Context) error {
	return merry.New("not implemented")
}

type BuildCommand struct {
	CompileCommand
}

func (c *BuildCommand) Action(*cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return merry.Wrap(err)
	}
	return c.CompileCommand.compile("build", "-o", wd)
}

type InstallCommand struct {
	CompileCommand
}

func (c *InstallCommand) Action(*cli.Context) error {
	return c.CompileCommand.compile("install")
}

type Commands struct {
	CommandBase

	Globals    *GlobalParams
	BuildCmd   *BuildCommand
	InstallCmd *InstallCommand
}

func Base(globals *GlobalParams) CommandBase {
	return CommandBase{
		Globals: globals,
	}
}

func NewCommands(passthroughArgs []string) *Commands {
	globals := &GlobalParams{
		PassthroughArgs: passthroughArgs,
	}
	return &Commands{
		CommandBase: Base(globals),
		BuildCmd: &BuildCommand{
			CompileCommand: CompileCommand{
				CommandBase: Base(globals),
			},
		},
		InstallCmd: &InstallCommand{
			CompileCommand: CompileCommand{
				CommandBase: Base(globals),
			},
		},
	}
}
