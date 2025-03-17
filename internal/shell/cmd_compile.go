package shell

import (
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
	DbOptions   cli.StringSlice
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
	genResult, err := c.generate(workingDir)
	if err != nil {
		return merry.Wrap(err)
	}

	if compileCmd == "" {
		return run.GoBin(workingDir, c.goBin, "mod", "tidy")
	}

	args := []string{compileCmd}
	args = append(args, compileArgs...)
	// -ldflags can be repeated so this doesn't interfere with PassthroughArgs
	ldflags := `-X github.com/xo/usql/text.CommandVersion=` + makeVersion(genResult.DownloadedUsqlVersion)
	args = append(args, "-ldflags", ldflags)

	// NB: This might interfere with PassthroughArgs
	// Required to avoid go mod tidy when adding just imports
	args = append(args, "-mod=mod")

	args = append(args, c.Globals.PassthroughArgs...)
	args = append(args, ".")
	return run.GoBin(workingDir, c.goBin, args...)
}

func makeVersion(downloadedVersion string) string {
	// we use _ as separator so it doesn't interfere with the suggested go install logic in usql/main.go
	return downloadedVersion + "_usqlgen"
}

func (c *CompileCommand) generate(workingDir string) (gen.Result, error) {
	genInput := gen.Input{
		Imports:     c.Imports.Value(),
		Replaces:    c.Replaces.Value(),
		Gets:        c.Gets.Value(),
		WorkingDir:  workingDir,
		USQLVersion: c.USQLVersion,
		USQLModule:  c.USQLModule,
	}
	err := applyOptionsFromNames(c.DbOptions.Value(), &genInput)
	if err != nil {
		return gen.Result{}, err
	}
	return c.generator(genInput)
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
		&cli.StringSliceFlag{
			Name:        "db-option",
			Usage:       `option that modifies configuration for newly imported drivers; use "usqlgen list options" to see what options are available`,
			Destination: &c.DbOptions,
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
