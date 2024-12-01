package shell

import (
	"github.com/ansel1/merry/v2"
	"github.com/sclgo/usqlgen/pkg/sclerr"
	"github.com/urfave/cli/v2"
	"io"
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

func (c *BuildCommand) Action(writer io.Writer) error {
	if writer == nil {
		writer = os.Stdout
	}
	
	destination := c.output
	if destination == "" {
		var err error
		destination, err = os.Getwd()
		if err != nil {
			return merry.Wrap(err)
		}
	}

	if c.output == "-" {
		// NB: Find a way to avoid creating another temp file
		tmpFile, err := os.CreateTemp("", "usqlgen")
		if err != nil {
			return merry.Wrap(err)
		}
		err = tmpFile.Close()
		if err != nil {
			return merry.Wrap(err)
		}
		defer func() {
			_ = os.Remove(tmpFile.Name())
		}()
		destination = tmpFile.Name()
	}

	err := c.CompileCommand.compile("build", "-o", destination)
	if err != nil {
		return err
	}

	if c.output == "-" {
		var destFile *os.File
		destFile, err = os.Open(destination)
		if err != nil {
			return merry.Wrap(err)
		}
		defer sclerr.CloseQuietly(destFile)
		_, err = io.Copy(writer, destFile)
	}

	return err

}
