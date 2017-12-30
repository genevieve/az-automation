package az

import (
	"bytes"
	"os/exec"
)

type CLI struct {
	path string
}

func NewCLI(path string) CLI {
	return CLI{
		path: path,
	}
}

func (c CLI) Execute(args []string) (string, error) {
	outBuffer := bytes.NewBuffer([]byte{})
	errBuffer := bytes.NewBuffer([]byte{})

	cmd := exec.Command(c.path, args...)
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer

	err := cmd.Run()
	if err != nil {
		return errBuffer.String(), err
	}

	return outBuffer.String(), nil
}
