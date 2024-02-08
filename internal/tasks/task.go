package tasks

import (
	"github.com/apudiu/server-backup/internal/server"
	"golang.org/x/crypto/ssh"
	"io"
)

type ServerTask interface {
	Execute(serverConn *ssh.Client) (result []byte, err error)
	ExecuteLive(serverConn *ssh.Client) (start, wait, closeFn func() error, err error)
}

type Task struct {
	Command   string
	StdOutErr io.Reader
	Succeeded bool
	ExecErr   error
}

func (t *Task) Execute(c *ssh.Client) (result []byte, err error) {
	result, err = server.ExecCmd(c, t.Command)
	return
}

func (t *Task) ExecuteLive(c *ssh.Client) (start, wait, closeFn func() error, err error) {
	start, wait, closeFn, err = server.ExecCmdLive(c, t.Command, &t.StdOutErr)
	return
}

func New(command string) *Task {
	t := &Task{
		Command: command,
	}
	return t
}
