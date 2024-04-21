package gen

import (
	"github.com/ansel1/merry/v2"
	"io/fs"
	"os"
	"path/filepath"
)

func chmod(path string) error {
	return merry.Wrap(filepath.WalkDir(path, chmodHandler))
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
