package shell

// Defines commands too short for their own file and the Commands object

import (
	"os"

	"github.com/ansel1/merry/v2"
	"github.com/urfave/cli/v2"
)

type InstallCommand struct {
	CompileCommand
}

func (c *InstallCommand) Action(*cli.Context) error {
	return c.CompileCommand.compile("install")
}

type GenerateCommand struct {
	CompileCommand

	output string
}

func (c *GenerateCommand) MakeFlags() []cli.Flag {
	return append(c.CompileCommand.MakeFlags(),
		&cli.StringFlag{
			Name:        "output",
			Usage:       `path to directory where the generated code will be written`,
			Aliases:     []string{"o"},
			Destination: &c.output,
			Value:       ".",
		})
}

func (c *GenerateCommand) Action(*cli.Context) error {
	err := os.MkdirAll(c.output, 0700)
	if err != nil {
		return merry.Wrap(err)
	}
	_, err = c.generate(c.output)
	return err
}

type Commands struct {
	CommandBase

	Globals     *GlobalParams
	BuildCmd    *BuildCommand
	InstallCmd  *InstallCommand
	GenerateCmd *GenerateCommand
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
			CompileCommand: MakeCompileCmd(globals),
		},
		InstallCmd: &InstallCommand{
			CompileCommand: MakeCompileCmd(globals),
		},
		GenerateCmd: &GenerateCommand{
			CompileCommand: MakeCompileCmd(globals),
		},
	}
}
