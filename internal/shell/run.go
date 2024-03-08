package shell

import (
	"log"
	"os"
	"slices"

	"github.com/urfave/cli/v2"
)

func RunArgs(args []string) {
	regularArgs, passthroughArgs := splitArgs(args)
	app := makeApp(passthroughArgs)
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
	RunArgs(os.Args)
}

func makeApp(passthroughArgs []string) *cli.App {
	commands := NewCommands(passthroughArgs)
	app := &cli.App{
		Name:  "scl",
		Usage: "Distribution generator for xo/usql - https://github.com/sclgo/usqlgen",
		Flags: commands.MakeFlags(),
		Args:  false,
		Commands: []*cli.Command{
			{
				Name:   "build",
				Usage:  "builds a usql binary distribution in the given directory",
				Args:   false,
				Flags:  commands.BuildCmd.MakeFlags(),
				Action: commands.BuildCmd.Action,
			},
			{
				Name:   "install",
				Usage:  "installs a usql binary distribution using 'go install'",
				Args:   false,
				Flags:  commands.InstallCmd.MakeFlags(),
				Action: commands.InstallCmd.Action,
			},
		},
	}
	return app
}
