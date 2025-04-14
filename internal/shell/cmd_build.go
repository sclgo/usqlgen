package shell

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ansel1/merry/v2"
	"github.com/murfffi/gorich/sclerr"
	"github.com/urfave/cli/v2"
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

// Action executes the build command using the given stdout
func (c *BuildCommand) Action(stdout io.Writer) error {
	if stdout == nil {
		stdout = os.Stdout
	}

	destination := c.output
	if destination == "" {
		destination = "." // will replaced by absolute path below
	}

	if c.output == "-" {
		// NB: Find a way to avoid creating another temp file
		var err error
		destination, err = touchTempFile()
		if err != nil {
			return err
		}
		defer func() {
			_ = os.Remove(destination)
		}()
	}

	destination, err := filepath.Abs(destination)
	if err != nil {
		return merry.Wrap(err)
	}

	err = c.CompileCommand.compile("build", "-o", destination)
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
		_, err = io.Copy(stdout, destFile)
	}

	return err

}

func touchTempFile() (string, error) {
	tmpFile, err := os.CreateTemp("", "usqlgen")
	if err != nil {
		return "", merry.Wrap(err)
	}
	err = tmpFile.Close()
	if err != nil {
		return "", merry.Wrap(err)
	}
	return tmpFile.Name(), nil
}
