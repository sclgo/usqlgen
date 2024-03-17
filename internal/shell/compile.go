package shell

import (
	"github.com/ansel1/merry/v2"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

type CompileCommand struct {
	CommandBase

	Imports cli.StringSlice
}

func (c *CompileCommand) compile(compileCmd string, compileArgs ...string) error {
	tmpDir, err := os.MkdirTemp("", "usqlgen")
	if err != nil {
		return merry.Wrap(err)
	}
	defer os.RemoveAll(tmpDir)
	workingDir := filepath.Join(tmpDir, "usql")
	err = os.Mkdir(workingDir, 0700)
	if err != nil {
		return merry.Wrap(err)
	}
	genInput := gen.Input{
		Imports:    c.Imports.Value(),
		WorkingDir: workingDir,
	}
	err = genInput.All()
	if err != nil {
		return merry.Wrap(err)
	}

	args := []string{compileCmd}
	args = append(args, compileArgs...)
	args = append(args, c.Globals.PassthroughArgs...)
	args = append(args, ".")
	return run.Go(workingDir, args...)
}

func (c *CompileCommand) MakeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "import",
			Destination: &c.Imports,
		},
	}
}
