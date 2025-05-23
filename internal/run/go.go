// Package run implements wrappers for running "go" command
package run

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ansel1/merry/v2"
)

func Go(workingDir string, goCmd ...string) error {
	return GoBin(workingDir, nil, FindGo(), goCmd...)
}

func FindGo() string {
	goBin := "go"
	goroot := os.Getenv("GOROOT")
	if goroot != "" {
		goBin = filepath.Join(goroot, "bin", goBin)
	}
	return goBin
}

// GoBin runs a go or a go-like command with a custom binary, capturing error output in the error result
func GoBin(workingDir string, addEnv []string, goBin string, goCmd ...string) error {
	log.Printf("running with addEnv %v %s %+v", addEnv, goBin, goCmd)
	cmd := exec.Command(goBin, goCmd...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	var buf bytes.Buffer
	cmd.Stderr = io.MultiWriter(&buf, os.Stderr)
	cmd.Env = append(os.Environ(), addEnv...)
	err := cmd.Run()
	return merry.Wrap(err, merry.AppendMessagef("while running %s %+v with output \n%s", goBin, goCmd, &buf))
}
