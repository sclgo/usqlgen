package gen

import (
	_ "embed"
	"github.com/ansel1/merry/v2"
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/sclgo/usqlgen/pkg/sclerr"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed main.go.tpl
var mainTpl string

//go:embed dbmgr.go
var dbMgrCode []byte

const fileMode = 0700

type Input struct {
	Imports    []string
	Replaces   []string
	WorkingDir string
}

func (i Input) Main(w io.Writer) error {
	tpl := template.Must(template.New("main").Parse(mainTpl))
	return merry.Wrap(tpl.Execute(w, i))
}

func (i Input) All() error {

	err := os.MkdirAll(i.WorkingDir, fileMode)
	if err != nil {
		return merry.Wrap(err)
	}

	err = i.runGo("mod", "init", "usql")
	if err != nil {
		return err
	}

	err = i.populateMain()
	if err != nil {
		return err
	}

	err = i.populateDbMgr()
	if err != nil {
		return err
	}

	// TODO Find a way to choose version of usql
	for _, rs := range append(
		[]string{"github.com/xo/usql=github.com/sclgo/usql@main"},
		i.Replaces...) {
		err = i.runGoModReplace(rs)
		if err != nil {
			return err
		}
		// go doesn't support running two go mod edit -replace without a
		// go mod tidy in between.
		err = i.runGo("mod", "tidy")
		if err != nil {
			return err
		}
	}

	return err
}

func (i Input) populateMain() error {
	mainFile, err := os.Create(filepath.Join(i.WorkingDir, "main.go"))
	if err != nil {
		return merry.Wrap(err)
	}
	defer sclerr.CloseQuietly(mainFile)

	err = i.Main(mainFile)
	if err != nil {
		return err
	}

	return merry.Wrap(mainFile.Close())
}

func (i Input) runGo(goCmd ...string) error {
	return run.Go(i.WorkingDir, goCmd...)
}

func (i Input) runGoModReplace(replaceSpec string) error {
	return i.runGo("mod", "edit", "-replace", replaceSpec)
}

func (i Input) populateDbMgr() error {
	genPackageDir := filepath.Join(i.WorkingDir, "gen")
	err := os.Mkdir(genPackageDir, fileMode)
	if err != nil {
		return merry.Wrap(err)
	}

	err = os.WriteFile(filepath.Join(genPackageDir, "dbmgr.go"), dbMgrCode, fileMode)
	return merry.Wrap(err)
}
