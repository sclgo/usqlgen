package shell

import (
	"io"
	"log"
	"os"
	"slices"

	"github.com/urfave/cli/v2"
)

func RunArgs(args []string, writer, errWriter io.Writer) {
	regularArgs, passthroughArgs := splitArgs(args)
	app := makeApp(passthroughArgs, writer, errWriter)
	if err := app.Run(regularArgs); err != nil {
		// logs merry errors better than panic
		log.Fatalf("%+v", err)
	}
}

func splitArgs(args []string) (regularArgs []string, passthroughArgs []string) {
	sepIndex := slices.Index(args, "--")
	if sepIndex == -1 || sepIndex == len(args)-1 {
		regularArgs = args
		return
	}

	regularArgs = args[:sepIndex]
	passthroughArgs = args[sepIndex+1:]
	return
}

func Run() {
	RunArgs(os.Args, nil, nil)
}

func makeApp(passthroughArgs []string, writer, errWriter io.Writer) *cli.App {
	commands := NewCommands(passthroughArgs)
	app := &cli.App{
		Description: "Distribution generator for xo/usql. Learn more at https://github.com/sclgo/usqlgen",
		Flags:       commands.MakeFlags(),
		Args:        false,
		Writer:      writer,
		ErrWriter:   errWriter,
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "builds a usql binary distribution in the given directory",
				Args:  false,
				Flags: commands.BuildCmd.MakeFlags(),
				Action: func(context *cli.Context) error {
					return commands.BuildCmd.Action(writer)
				},
			},
			{
				Name:   "install",
				Usage:  "installs a usql binary distribution using 'go install'",
				Args:   false,
				Flags:  commands.InstallCmd.MakeFlags(),
				Action: commands.InstallCmd.Action,
			},
			{
				Name:   "generate",
				Usage:  "generates the code for the usql binary distribution without compiling it",
				Args:   false,
				Flags:  commands.GenerateCmd.MakeFlags(),
				Action: commands.GenerateCmd.Action,
			},
			{
				Name:  "list",
				Usage: "subcommands list various options and attributes",
				Subcommands: []*cli.Command{
					{
						Name:   "options",
						Usage:  "displays options available for --dboptions parameter. Each option modifies how database drivers are treated.",
						Args:   false,
						Action: listOptions,
					},
				},
			},
		},
	}
	app.Usage = app.Description
	return app
}
