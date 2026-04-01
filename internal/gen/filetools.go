package gen

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ansel1/merry/v2"
)

func chmod(path string) error {
	return merry.Wrap(filepath.WalkDir(path, chmodHandler))
}

func chmodHandler(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		return os.Chmod(path, fileMode)
	}
	return nil
}
