package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/server"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	//generateEmptyConfigFile()

	c := config.Config{}
	c.Parse()
	//fmt.Printf("%+v \n", c)

	for _, server := range c.Servers {
		doWork(&server)
	}

}

func doWork(s *config.ServerConfig) {
	//stdOut, stdErr, err := remoteRun(s, "ls /var/www/php80")
	//fmt.Println("stdOut", stdOut)
	//fmt.Println("stdErr", stdErr)
	//fmt.Println("err", err)

	conn, connErr := server.ConnectToServer(s)
	util.FailIfErr(connErr, "Connection establishment with server "+s.Ip.String()+" failed")
	defer conn.Close()

	// task
	zipProject(conn)
}

func zipProject(c *ssh.Client) (bool, error) {

	//t := tasks.Task{
	//	Commands:  nil,
	//	StdOut:    nil,
	//	StdErr:    nil,
	//	Succeeded: false,
	//	ExecErr:   nil,
	//}

	cmdList := []string{
		"find /var -name 'ven*'",
		"ls -la /var/www",
	}

	for _, cmd := range cmdList {
		fmt.Println("CMD: ", cmd)

		stdOut, stdErr, cmdErr := server.ExecCmd(c, cmd)
		if isEOFErr := errors.Is(cmdErr, io.EOF); !isEOFErr {
			fmt.Println("cmdErr", cmdErr)
			return false, cmdErr
		}

		bb := bytes.NewBuffer(nil)

		io.Copy(bb, stdOut)
		io.Copy(os.Stderr, stdErr)

		fmt.Println("from buff", bb.String())
	}

	return true, nil
}

// e.g. output, err := remoteRun("root", "MY_IP", "PRIVATE_KEY", "ls")
func remoteRun(c *config.ServerConfig, cmd string) (string, string, error) {
	// privateKey could be read from a file, or retrieved from another storage
	// source, such as the Secret Service / GNOME Keyring
	key, err := ssh.ParsePrivateKey(util.ReadFromFile(c.Key))
	if err != nil {
		return "", "", err
	}
	// Authentication
	cfg := &ssh.ClientConfig{
		User: c.User,
		// https://github.com/golang/go/issues/19767
		// as clientConfig is non-permissive by default
		// you can set ssh.InsercureIgnoreHostKey to allow any host
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
			ssh.Password(c.Password),
		},
		//alternatively, you could use a password
		/*
		   Auth: []ssh.AuthMethod{
		       ssh.Password("PASSWORD"),
		   },
		*/
	}
	// Connect
	client, err := ssh.Dial("tcp", net.JoinHostPort(c.Ip.String(), strconv.Itoa(c.Port)), cfg)
	if err != nil {
		return "", "", err
	}
	// Create a session. It is one session per command.
	session, err := client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	var stdOut bytes.Buffer
	session.Stdout = &stdOut // get output
	// you can also pass what gets input to the stdin, allowing you to pipe
	// content from client to server
	//      session.Stdin = bytes.NewBufferString("My input")

	var stdErr bytes.Buffer
	session.Stderr = &stdErr

	// Finally, run the command
	err = session.Run(cmd)
	return stdOut.String(), stdErr.String(), err
}
