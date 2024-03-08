package shell

import (
	"github.com/ansel1/merry/v2"
	"github.com/urfave/cli/v2"
)

type GlobalParams struct {
	Verbose         bool
	PassthroughArgs []string
}

type CommandBase struct {
	Globals *GlobalParams
}

func (CommandBase) MakeFlags() []cli.Flag {
	return nil
}

func (CommandBase) Action(*cli.Context) error {
	return merry.New("not implemented")
}

type BuildCommand struct {
	CommandBase
}

type InstallCommand struct {
	CommandBase
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
			CommandBase: Base(globals),
		},
		InstallCmd: &InstallCommand{
			CommandBase: Base(globals),
		},
	}
}
