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
			Usage:       "enabled detailed, debug output",
			Aliases:     []string{"v"},
			Destination: &c.Globals.Verbose,
		},
	}
}

func (*CommandBase) Action(*cli.Context) error {
	return merry.New("not implemented")
}

type BuildCommand struct {
	CompileCommand

	output string
}

func (c *BuildCommand) MakeFlags() []cli.Flag {
	return append(c.CompileCommand.MakeFlags(),
		&cli.StringFlag{
			Name: "output",
			Usage: `path to file where the executable will be written; 
if value is -, binary will be written to standard output; if the path is a directory,
executable with name 'usql' will be created there;`,
			Aliases:     []string{"o"},
			Destination: &c.output,
			Value:       ".",
		})
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

type GenerateCommand struct {
	CompileCommand

	output string
}

func (c *GenerateCommand) MakeFlags() []cli.Flag {
	return append(c.CompileCommand.MakeFlags(),
		&cli.StringFlag{
			Name: "output",
			Usage: `path to directory where the generated code will be written; 
if value is -, tar archive will be written to standard output`,
			Aliases:     []string{"o"},
			Destination: &c.output,
			Value:       ".",
		})
}

func (c *GenerateCommand) Action(*cli.Context) error {
	return c.CompileCommand.compile("")
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

func MakeCompileCmd(globals *GlobalParams) CompileCommand {
	return CompileCommand{
		CommandBase: Base(globals),
	}
}
