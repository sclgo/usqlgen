package run

import (
	"bytes"
	"github.com/ansel1/merry/v2"
	"io"
	"log"
	"os"
	"os/exec"
)

func Go(workingDir string, goCmd ...string) error {
	log.Printf("running go %+v", goCmd)
	cmd := exec.Command("go", goCmd...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	var buf bytes.Buffer
	cmd.Stderr = io.MultiWriter(&buf, os.Stderr)
	err := cmd.Run()
	return merry.Wrap(err, merry.AppendMessagef("while running go %+v with output \n%s", goCmd, &buf))
}
