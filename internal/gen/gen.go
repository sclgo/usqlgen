package gen

import (
	_ "embed"
	"github.com/ansel1/merry/v2"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

//go:embed main.go.tpl
var mainTpl string

type Input struct {
	Imports    []string
	WorkingDir string
}

func (i Input) Main(w io.Writer) error {
	tpl := template.Must(template.New("main").Parse(mainTpl))
	return tpl.Execute(w, i)
}

func (i Input) All() error {
	err := os.MkdirAll(i.WorkingDir, 0700)
	if err != nil {
		return merry.Wrap(err)
	}

	err = i.runGo("mod", "init", "usql")
	if err != nil {
		return merry.Wrap(err)
	}
	mainFile, err := os.Create(filepath.Join(i.WorkingDir, "main.go"))
	if err != nil {
		return merry.Wrap(err)
	}

	err = i.Main(mainFile)
	if err != nil {
		return merry.Wrap(err)
	}

	err = i.runGo("mod", "edit", "-replace", "github.com/xo/usql=github.com/sclgo/usql@latest")
	if err != nil {
		return err
	}

	err = i.runGo("mod", "tidy")
	return err
}

func (i Input) runGo(goCmd ...string) error {
	cmd := exec.Command("go", goCmd...)
	cmd.Dir = i.WorkingDir
	err := cmd.Run()
	return merry.Wrap(err, merry.AppendMessagef("while running %+v", goCmd))
}
