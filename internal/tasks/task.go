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
	StdOut, StdErr io.Writer
	Succeeded      bool
	ExecErr        error
}

func (t *Task) Execute(c *ssh.Client) error {

	_, _, err := server.ExecCmd(c, t.Commands[0])
	if isEOFErr := errors.Is(err, io.EOF); !isEOFErr && err != nil {
		return err
	}
	fmt.Println("after exec")

	//nw := bufio.NewWriter(stdOut)

	//for {
	//	n, err2 := stdOut.Write()
	//	if err2 != nil {
	//		fmt.Printf("Read %d bytes\n", n)
	//		fmt.Println("Copy err:", err.Error())
	//		break
	//	}
	//	fmt.Printf(".")
	//}

	//for _, cmd := range t.Commands {
	//	stdOut, stdErr, err := server.ExecCmd(c, cmd)
	//	if isEOFErr := errors.Is(err, io.EOF); !isEOFErr {
	//		return err
	//	}
	//
	//	t.Succeeded, t.StdOut, t.StdErr, t.ExecErr = true, stdOut, stdErr, err
	//}
	return nil
}

func New(commands []string) *Task {
	t := &Task{
		Commands: commands,
	}
	return t
}
