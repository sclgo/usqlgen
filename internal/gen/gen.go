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
	"github.com/murfffi/gorich/lang"
	"github.com/murfffi/gorich/sclerr"
	"github.com/sclgo/usqlgen/internal/run"
	"modernc.org/fileutil"
)

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
	KeepCgo          bool
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
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 || outputBuf.Len() > 0 {
			return result, merry.Wrap(err, merry.AppendMessagef("while running go mod download with stdout length %d and stderr output \n%s", outputBuf.Len(), &errorBuf))
		}
		// We ignore exit code 1 with non-empty output, because this indicates a partial success of
		// go mod download command and that the package was likely successfully downloaded.
		// This case happens frequently.
		// https://github.com/golang/go/issues/35380 is about a different issue but some comments
		// cover this case.
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

	if i.Imports != nil {
		err = i.replaceMain()
		if err != nil {
			return result, err
		}
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

	if !i.KeepCgo {
		adjustErr := i.adjustCgoTags()
		if adjustErr != nil {
			i.log("Failed to adjust base cgo tags, but this might not be an issue, depending on tags and environment. Cause: %v", adjustErr)
		}
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

func (i Input) replaceInUsqlFile(relPath string, before string, after string) error {
	origMainFile := filepath.Join(i.WorkingDir, relPath)
	origMainBytes, err := os.ReadFile(origMainFile)
	if err != nil {
		return merry.Wrap(err)
	}
	origMainBytes = bytes.Replace(origMainBytes, []byte(before), []byte(after), 1)
	err = os.WriteFile(origMainFile, origMainBytes, fileMode)
	return merry.Wrap(err)
}

func (i Input) replaceMain() error {
	err := i.replaceInUsqlFile("main.go", "func main()", "func origMain()")
	if err != nil {
		return err
	}

	return i.populateMain()
}

func (i Input) adjustCgoTags() error {
	err := i.replaceInUsqlFile(filepath.Join("internal", "sqlite3.go"), "!no_base", "!no_base && cgo")
	if err != nil {
		return err
	}
	moderncsqliteHookPath := filepath.Join("internal", "moderncsqlite.go")
	// we must include moderncsqlite *only* if sqlite3 was excluded because of !go
	return i.replaceInUsqlFile(moderncsqliteHookPath, "most", "most || (!cgo && !no_base && !no_sqlite3)")

	// usql already contains code that assigns sqlite3 aliases to moderncsqlite,
	// if moderncsqlite is present but sqlite3 is not - usql/internal/z.go
}

func (i Input) log(msg string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, msg, args...)
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
