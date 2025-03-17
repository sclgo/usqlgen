package gen

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/ansel1/merry/v2"
	"github.com/sclgo/usqlgen/internal/run"
	"github.com/sclgo/usqlgen/pkg/lang"
	"github.com/sclgo/usqlgen/pkg/sclerr"
	"modernc.org/fileutil"
)

//go:embed main.go.tpl
var mainTpl string

//go:embed dbmgr.go
var dbMgrCode []byte

const fileMode = 0700

type Input struct {

	// Imports lists packages that will be imported with blank import
	// in the order given. Duplicates are allowed and won't be removed.
	// Caller should remove duplicates if needed.
	Imports []string

	Replaces    []string
	Gets        []string
	WorkingDir  string
	USQLModule  string
	USQLVersion string

	IncludeSemicolon bool
}

type Result struct {
	DownloadedUsqlVersion string
}

func (i Input) Main(w io.Writer) error {
	tpl := template.Must(template.New("main").Parse(mainTpl))
	return merry.Wrap(tpl.Execute(w, i))
}

// AllDownload generates all usql distribution code using the go mod download strategy
func (i Input) AllDownload() (Result, error) {
	var result Result
	err := os.MkdirAll(i.WorkingDir, fileMode)
	if err != nil {
		return result, merry.Wrap(err)
	}

	cmd := exec.Command("go", "mod", "download", "-json", i.getUSQLModuleVersion())
	cmd.Dir = i.WorkingDir
	var outputBuf, errorBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = io.MultiWriter(&errorBuf, os.Stderr)
	err = cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return result, merry.Wrap(err, merry.AppendMessagef("while running go mod download with output \n%s", &errorBuf))
		}
		// https://github.com/golang/go/issues/35380
	}

	var downloadInfo map[string]any
	err = json.NewDecoder(&outputBuf).Decode(&downloadInfo)
	if err != nil {
		return result, merry.Wrap(err)
	}

	err = i.copyOriginal(downloadInfo)
	if err != nil {
		return result, err
	}

	result.DownloadedUsqlVersion = i.USQLVersion
	if downloadedVersion, ok := downloadInfo["Version"]; ok {
		result.DownloadedUsqlVersion = fmt.Sprint(downloadedVersion)
	}

	// We expect that the DownloadedUsqlVersion is already at go1.23+
	//err = i.runGo("mod", "edit", "-go=1.23")
	//if err != nil {
	//	return result, merry.Wrap(err)
	//}

	origMainFile := filepath.Join(i.WorkingDir, "main.go")
	origMainBytes, err := os.ReadFile(origMainFile)
	if err != nil {
		return result, merry.Wrap(err)
	}
	origMainBytes = bytes.Replace(origMainBytes, []byte("func main()"), []byte("func origMain()"), 1)
	err = os.WriteFile(origMainFile, origMainBytes, fileMode)
	if err != nil {
		return result, merry.Wrap(err)
	}

	err = i.populateMain()
	if err != nil {
		return result, err
	}

	err = i.populateDbMgr()
	if err != nil {
		return result, err
	}

	err = i.doGoGet()
	if err != nil {
		return result, err
	}

	err = i.goModReplace(i.Replaces)
	if err != nil {
		return result, err
	}

	return result, err
}

func (i Input) getUSQLModuleVersion() string {
	usqlModule := lang.IfEmpty(i.USQLModule, "github.com/xo/usql")
	usqlVersion := lang.IfEmpty(i.USQLVersion, "latest")
	return fmt.Sprintf("%s@%s", usqlModule, usqlVersion)
}

func (i Input) copyOriginal(downloadInfo map[string]any) error {
	if _, ok := downloadInfo["Dir"]; ok {
		return i.copyOriginalFromDir(downloadInfo)
	}
	if _, ok := downloadInfo["Zip"]; ok {
		return i.copyOriginalFromZip(downloadInfo)
	}
	return merry.Wrap(fmt.Errorf("can't copy original code; neither Zip or Dir available in go mod download output. Error field: %v", downloadInfo["Error"]))
}

func (i Input) copyOriginalFromZip(downloadInfo map[string]any) error {
	downloadZipAny, ok := downloadInfo["Zip"]
	if !ok {
		return merry.New("missing 'Zip' field")
	}

	downloadZip := downloadZipAny.(string)
	unpackedZip := filepath.Join(i.WorkingDir, "zip")
	err := ExtractZip(downloadZip, unpackedZip)
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
		return merry.Wrap(fmt.Errorf("failed to download module: %v", errorMsg))
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

	err = chmod(i.WorkingDir)
	return err
}

func (i Input) All() error {
	_, err := i.AllDownload()
	return err
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
	mainFile, err := os.Create(filepath.Join(i.WorkingDir, "new_main.go"))
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

func (i Input) doGoGet() error {
	return doGoGet(i.Gets, i)
}

// We believe that we don't need to go get the --imports params,
// since this is handled by CompileCmd using either -mod=mod for build/install,
// or "go mod tidy for generate
//
//func (i Input) getImports() error {
//	if i.Gets != nil {
//		// The modules in get may have already delivered those packages
//		// We don't want to run 'go get' for them because it will upgrade the explicit version
//		i.Imports = lo.Filter(i.Imports, func(item string, index int) bool {
//			return i.runGo("list", "-find", item) != nil // returns exit code 1 if not found
//		})
//	}
//	return doGoGet(i.Imports, i)
//}

func doGoGet(gets []string, i Input) error {
	for _, gs := range gets {
		err := i.runGo("get", gs)
		if err != nil {
			return err
		}
	}
	return nil
}
