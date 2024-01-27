package tasks

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/server"
	"golang.org/x/crypto/ssh"
	"io"
)

type ServerTask interface {
	Execute(serverConn *ssh.Client) (err error)
}

type Task struct {
	Commands  []string
	StdOutErr io.Reader
	Succeeded bool
	ExecErr   error
}

//func (t *Task) Execute(c *ssh.Client) error {
//	start, wait, closeFn, err := server.ExecCmdLive(c, t.Commands[0], &t.StdOutErr)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		err = closeFn()
//	}()
//
//	// read from task
//	ch := make(chan struct{})
//	l := logger.Logger{}
//
//	// read output while the cmd is executing
//	go func() {
//		l.ReadStream(&t.StdOutErr)
//		ch <- struct{}{}
//	}()
//
//	err = start()
//	<-ch
//
//	err = wait()
//	if err != nil {
//		fmt.Println("Wait err", err.Error())
//	}
//
//	l.LogToFile("zip.log")
//
//	fmt.Println("after exec")
//
//	//for _, cmd := range t.Commands {
//	//	stdOut, stdErr, err := server.ExecCmd(c, cmd)
//	//	if isEOFErr := errors.Is(err, io.EOF); !isEOFErr {
//	//		return err
//	//	}
//	//
//	//	t.Succeeded, t.StdOut, t.StdErr, t.ExecErr = true, stdOut, stdErr, err
//	//}
//	return nil
//}

func (t *Task) Execute(c *ssh.Client) error {
	res, err := server.ExecCmd(c, t.Commands[0])
	if err != nil {
		return err
	}

	fmt.Println("after exec \n", string(res))

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
