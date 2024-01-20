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
	"os"
)

func main() {
	//generateEmptyConfigFile()

	c := config.Config{}
	c.Parse()
	//fmt.Printf("%+v \n", c)

	for _, srv := range c.Servers {
		doWork(&srv)
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

	isExist, _ := server.RemoteIsPathExist(c, "/var/www/html/index.php")
	if isExist {
		fmt.Println("path exist")
	}

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
