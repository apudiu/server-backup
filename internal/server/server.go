package server

import (
	"errors"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"strconv"
)

func ConnectToServer(c *config.ServerConfig) (conn *ssh.Client, err error) {
	serverKey := util.ReadFromFile(c.Key)
	signer, err := ssh.ParsePrivateKey(serverKey)
	if err != nil {

		// todo: check for this err type
		//if errors.Is(err, ssh.PassphraseMissingError{}) {}

		// if failed try to parse key using the password
		signer, err = ssh.ParsePrivateKeyWithPassphrase(serverKey, []byte(c.Password))
		util.FailIfErr(err, "Server private key parsing failed")
	}

	//hostKeyCallback, hostKeyErr := knownhosts.New("~/.ssh/known_hosts")
	//failIfErr(hostKeyErr)

	conf := &ssh.ClientConfig{
		User: c.User,
		//HostKeyCallback: hostKeyCallback,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
			ssh.Password(c.Password),
		},
		//Timeout: 15 * time.Second,
	}

	// connect to server

	hostWithPort := net.JoinHostPort(c.Ip.String(), strconv.Itoa(c.Port))
	conn, err = ssh.Dial("tcp", hostWithPort, conf)
	return
}

func ExecCmd(conn *ssh.Client, cmd string) (stdOut, stdErr io.Reader, err error) {

	session, err := conn.NewSession()
	if err != nil {
		return
	}
	defer func() {
		er := session.Close()
		if !errors.Is(er, io.EOF) {
			err = er
		}
	}()

	// process CMD
	stdOut, err = session.StdoutPipe()
	if err != nil {
		return
	}

	stdErr, err = session.StderrPipe()
	if err != nil {
		return
	}

	//err = session.Wait()
	err = session.Run(cmd)
	return
}
