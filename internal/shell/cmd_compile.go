package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ansel1/merry/v2"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/urfave/cli/v2"
)

type CompileCommand struct {
	CommandBase
	generator func(gen.Input) (gen.Result, error)
	goBin     string

	Imports     cli.StringSlice
	Replaces    cli.StringSlice
	Gets        cli.StringSlice
	USQLModule  string
	USQLVersion string
}

func (c *CompileCommand) compile(compileCmd string, compileArgs ...string) error {
	tmpDir, err := os.MkdirTemp("", "usqlgen")
	if err != nil {
		return merry.Wrap(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	workingDir := filepath.Join(tmpDir, "usql")
	err = os.Mkdir(workingDir, 0700)
	if err != nil {
		return merry.Wrap(err)
	}
	genInput := gen.Input{
		Imports:     c.Imports.Value(),
		Replaces:    c.Replaces.Value(),
		Gets:        c.Gets.Value(),
		WorkingDir:  workingDir,
		USQLVersion: c.USQLVersion,
		USQLModule:  c.USQLModule,
	}
	genResult, err := c.generator(genInput)
	if err != nil {
		return merry.Wrap(err)
	}

	if compileCmd == "" {
		return nil
	}

	args := []string{compileCmd}
	args = append(args, compileArgs...)
	// -ldflags can be repeated so this doesn't interfere with PassthroughArgs
	ldflags := fmt.Sprintf(`-X github.com/xo/usql/text.CommandVersion=%s-usqlgen`, genResult.DownloadedUsqlVersion)
	args = append(args, "-ldflags", ldflags)
	args = append(args, c.Globals.PassthroughArgs...)
	args = append(args, ".")
	return run.GoBin(workingDir, c.goBin, args...)
}

func (c *CompileCommand) MakeFlags() []cli.Flag {
	return append([]cli.Flag{
		&cli.StringSliceFlag{
			Name:        "import",
			Usage:       "imports for side-effects the given package, typically for registering database/sql drivers, can be repeated",
			Aliases:     []string{"i"},
			Destination: &c.Imports,
		},
		&cli.StringSliceFlag{
			Name:        "replace",
			Usage:       "adds a replace directive to the generated module with the same format as 'go mod edit -replace', can be repeated",
			Aliases:     []string{"r"},
			Destination: &c.Replaces,
		},
		&cli.StringSliceFlag{
			Name:        "get",
			Usage:       "adds or updates the provided module using go get",
			Destination: &c.Gets,
		},
		&cli.StringFlag{
			Name:        "usql-module",
			Usage:       "module name of usql fork to use if needed",
			Value:       "github.com/xo/usql",
			Destination: &c.USQLModule,
		},
		&cli.StringFlag{
			Name:        "usql-version",
			Usage:       "usql version to use; can be any valid module version incl. 'latest', release, tag, branch, or Git commit",
			Aliases:     []string{"uv"},
			Value:       "latest",
			Destination: &c.USQLVersion,
		},
	}, c.CommandBase.MakeFlags()...)
}

func MakeCompileCmd(globals *GlobalParams) CompileCommand {
	return CompileCommand{
		generator:   gen.Input.AllDownload,
		goBin:       run.FindGo(),
		CommandBase: Base(globals),
	}
}
