package main

import (
	"bytes"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"net"
	"strconv"
)

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
