package tasks

import (
	"errors"
	"fmt"
	"github.com/apudiu/server-backup/internal/server"
	"golang.org/x/crypto/ssh"
	"io"
)

type ServerTask interface {
	Execute(serverConn *ssh.Client) (err error)
}

type Task struct {
	Commands       []string
	StdOut, StdErr io.Reader
	Succeeded      bool
	ExecErr        error
}

func (t *Task) Execute(c *ssh.Client) error {

	for _, cmd := range t.Commands {
		fmt.Println("CMD: ", cmd)

		stdOut, stdErr, err := server.ExecCmd(c, cmd)
		if isEOFErr := errors.Is(err, io.EOF); !isEOFErr {
			return err
		}

		t.Succeeded, t.StdOut, t.StdErr, t.ExecErr = true, stdOut, stdErr, err
	}
	return nil
}

func New(commands []string) *Task {
	t := &Task{
		Commands: commands,
	}
	return t
}
