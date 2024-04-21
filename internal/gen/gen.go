package gen

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ansel1/merry/v2"
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/sclgo/usqlgen/pkg/sclerr"
	"github.com/xyproto/unzip"
	"io"
	"io/fs"
	"modernc.org/fileutil"
	"os"
	"os/exec"
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

// AllDownload generates a all usql distribution code using the go mod download strategy
func (i Input) AllDownload() error {
	err := os.MkdirAll(i.WorkingDir, fileMode)
	if err != nil {
		return merry.Wrap(err)
	}

	// TODO Make version a parameter
	cmd := exec.Command("go", "mod", "download", "-json", "github.com/xo/usql@latest")
	cmd.Dir = i.WorkingDir
	var outputBuf, errorBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = io.MultiWriter(&errorBuf, os.Stderr)
	err = cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return merry.Wrap(err, merry.AppendMessagef("while running go mod download with output \n%s", &errorBuf))
		}
		// https://github.com/golang/go/issues/35380
	}

	var downloadInfo map[string]any
	err = json.NewDecoder(&outputBuf).Decode(&downloadInfo)
	if err != nil {
		return merry.Wrap(err)
	}

	err = i.copyOriginal(downloadInfo)
	if err != nil {
		return err
	}

	err = i.runGo("mod", "edit", "-go=1.22.0")
	if err != nil {
		return merry.Wrap(err)
	}

	origMainFile := filepath.Join(i.WorkingDir, "orig_main.go")
	err = os.Rename(filepath.Join(i.WorkingDir, "main.go"), origMainFile)
	if err != nil {
		return merry.Wrap(err)
	}

	origMainBytes, err := os.ReadFile(origMainFile)
	if err != nil {
		return merry.Wrap(err)
	}
	origMainBytes = bytes.Replace(origMainBytes, []byte("func main()"), []byte("func origMain()"), 1)
	err = os.WriteFile(origMainFile, origMainBytes, fileMode)
	if err != nil {
		return merry.Wrap(err)
	}

	err = i.populateMain()
	if err != nil {
		return err
	}

	err = i.populateDbMgr()
	if err != nil {
		return err
	}

	err = i.goModReplace(i.Replaces)
	if err != nil {
		return err
	}

	if len(i.Replaces) == 0 {
		err = i.runGo("mod", "tidy")
	}

	if err != nil {
		return err
	}

	return err
}

func (i Input) copyOriginal(downloadInfo map[string]any) error {
	if _, ok := downloadInfo["Dir"]; ok {
		return i.copyOriginalFromDir(downloadInfo)
	}
	if _, ok := downloadInfo["Zip"]; ok {
		return i.copyOriginalFromZip(downloadInfo)
	}
	return merry.Wrap(fmt.Errorf("neither Zip or Dir available in go mod download output. Error field: %v", downloadInfo["Error"]))
}

func (i Input) copyOriginalFromZip(downloadInfo map[string]any) error {
	downloadZipAny, ok := downloadInfo["Zip"]
	if !ok {
		return merry.New("missing 'Zip' field")
	}

	downloadZip := downloadZipAny.(string)
	unpackedZip := filepath.Join(i.WorkingDir, "zip")
	err := unzip.Extract(downloadZip, unpackedZip)
	if err != nil {
		return merry.Wrap(err)
	}

	// Go to folder with actual code
	var entries []os.DirEntry
	for entries, err = os.ReadDir(unpackedZip); len(entries) == 1 && err == nil; {
		unpackedZip = filepath.Join(unpackedZip, entries[0].Name())
		entries, err = os.ReadDir(unpackedZip)
	}
	if err != nil {
		return merry.Wrap(err)
	}

	_, _, err = fileutil.CopyDir(os.DirFS(unpackedZip), i.WorkingDir, ".", nil)
	if err != nil {
		return merry.Wrap(err)
	}

	return nil
}

func (i Input) copyOriginalFromDir(downloadInfo map[string]any) error {
	errorMsg, ok := downloadInfo["Error"]
	if ok {
		return merry.Errorf("Failed to download module: %v", errorMsg)
	}

	downloadDirAny, ok := downloadInfo["Dir"]
	if !ok {
		return merry.New("missing 'Dir' field")
	}
	downloadDir := downloadDirAny.(string)
	_, _, err := fileutil.CopyDir(os.DirFS(downloadDir), i.WorkingDir, ".", nil)
	if err != nil {
		return merry.Wrap(err)
	}

	err = filepath.WalkDir(i.WorkingDir, chmodHandler)
	if err != nil {
		return merry.Wrap(err)
	}
	return nil
}

//nolint:unused seems to be a bug in staticheck U1000 but couldn't reproduce in a minimal example
func chmodHandler(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		return os.Chmod(path, fileMode)
	}
	return nil
}

func (i Input) All() error {
	return i.AllDownload()
}

// AllFork generates all usql distribution code using the fork strategy
func (i Input) AllFork() error {

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

	// There seems to be no way to choose usql version with this strategy.
	// See also AllDownload.
	err = i.goModReplace(append(
		[]string{"github.com/xo/usql=github.com/sclgo/usql@main"},
		i.Replaces...))
	if err != nil {
		return err
	}

	return err
}

func (i Input) goModReplace(replaceList []string) error {
	for _, rs := range replaceList {

		// Consider golang.org/x/mod/modfile
		err := i.runGoModReplace(rs)
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
	return nil
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
	err := os.MkdirAll(genPackageDir, fileMode)
	if err != nil {
		return merry.Wrap(err)
	}

	err = os.WriteFile(filepath.Join(genPackageDir, "dbmgr.go"), dbMgrCode, fileMode)
	return merry.Wrap(err)
}
