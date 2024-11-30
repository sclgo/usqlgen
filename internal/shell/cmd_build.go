package shell

import (
	"github.com/ansel1/merry/v2"
	"github.com/urfave/cli/v2"
	"os"
)

type BuildCommand struct {
	CompileCommand

	output string
}

func (c *BuildCommand) MakeFlags() []cli.Flag {
	return append(c.CompileCommand.MakeFlags(),
		&cli.StringFlag{
			Name: "output",
			Usage: `path to file or directory where the executable will be written; 
if value is -, binary will be written to standard output; if the path is a directory,
executable with name 'usql' will be created there;`,
			Aliases:     []string{"o"},
			Destination: &c.output,
			Value:       ".",
		})
}

func (c *BuildCommand) Action(*cli.Context) error {
	destination, err := os.Getwd()
	if err != nil {
		return merry.Wrap(err)
	}

	// TODO Use c.output
	return c.CompileCommand.compile("build", "-o", destination)
}
