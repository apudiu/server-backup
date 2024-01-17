package main

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	//generateEmptyConfigFile()

	c := config{}
	c.Parse()
	//fmt.Printf("%+v \n", c)

	for _, server := range c.Servers {
		doWork(&server)
	}

}

func doWork(s *serverConfig) {
	//stdOut, stdErr, err := remoteRun(s, "ls /var/www/php80")
	//fmt.Println("stdOut", stdOut)
	//fmt.Println("stdErr", stdErr)
	//fmt.Println("err", err)

	connectToServer(s)
}

func connectToServer(c *serverConfig) {
	serverKey := readFromFile(c.Key)

	signer, err := ssh.ParsePrivateKey(serverKey)
	if err != nil {

		// todo: check for this err type
		//if errors.Is(err, ssh.PassphraseMissingError{}) {}

		// if failed try to parse key using the password
		signer, err = ssh.ParsePrivateKeyWithPassphrase(serverKey, []byte(c.Password))
		failIfErr(err, "Server private key parsing failed")
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
	conn, err := ssh.Dial("tcp", hostWithPort, conf)
	failIfErr(err, "Connection establishment with server "+c.Ip.String()+" failed")
	defer conn.Close()

	cmds := []string{
		"find /var -name '*ven*'",
		"ls -la /var/www",
	}

	for _, cmd := range cmds {
		stdOut, stdErr, cmdErr := execCmd(conn, cmd)
		if cmdErr != nil {
			fmt.Println(cmdErr)
		}
		fmt.Println("CMD: ", cmd)

		io.Copy(os.Stdout, stdOut)
		io.Copy(os.Stderr, stdErr)
	}

}

func execCmd(conn *ssh.Client, cmd string) (stdOut, stdErr io.Reader, err error) {

	session, err := conn.NewSession()
	failIfErr(err, "Getting new session error on server")
	defer session.Close()

	// process CMD
	stdOut, err = session.StdoutPipe()
	if err != nil {
		return
	}

	stdErr, err = session.StderrPipe()
	if err != nil {
		return
	}

	//session.Run("find /var -name '*ven*'")
	session.Run(cmd)
	return
}

// e.g. output, err := remoteRun("root", "MY_IP", "PRIVATE_KEY", "ls")
func remoteRun(c *serverConfig, cmd string) (string, string, error) {
	// privateKey could be read from a file, or retrieved from another storage
	// source, such as the Secret Service / GNOME Keyring
	key, err := ssh.ParsePrivateKey(readFromFile(c.Key))
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
