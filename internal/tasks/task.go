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

//func (t *Task) ExecuteLive(c *ssh.Client) error {
//	start, wait, closeFn, err := server.ExecCmdLive(c, t.Command[0], &t.StdOutErr)
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
//	l.WriteToFile("zip.log")
//
//	fmt.Println("after exec")
//
//	//for _, cmd := range t.Command {
//	//	stdOut, stdErr, err := server.ExecCmd(c, cmd)
//	//	if isEOFErr := errors.Is(err, io.EOF); !isEOFErr {
//	//		return err
//	//	}
//	//
//	//	t.Succeeded, t.StdOut, t.StdErr, t.ExecErr = true, stdOut, stdErr, err
//	//}
//	return nil
//}

func New(command string) *Task {
	t := &Task{
		Command: command,
	}
	return t
}
