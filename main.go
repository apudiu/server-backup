package main

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	generateEmptyConfigFile()

	//c := config{}
	//c.Parse()
	////fmt.Printf("%#v", c)
	//
	//// login to servers
	//s := c.Servers[0]
	//
	////connectToServer(&s)
	//res, err := remoteRun(&s, "ls /var/www/php80")
	//fmt.Println(res, err)
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

	hostKeyCallback, hostKeyErr := knownhosts.New("~/.ssh/known_hosts")
	failIfErr(hostKeyErr)

	conf := &ssh.ClientConfig{
		User:            c.User,
		HostKeyCallback: hostKeyCallback,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
			ssh.Password(c.Password),
		},
		//Timeout: 15 * time.Second,
	}

	// connect to server

	conn, err := ssh.Dial("tcp", c.Ip.String(), conf)
	failIfErr(err, "Connection establishment with server "+c.Ip.String()+" failed")
	defer conn.Close()

	session, err := conn.NewSession()
	failIfErr(err, "Getting new session error on server "+c.Ip.String())
	defer session.Close()

	// process CMD

	var stdin io.WriteCloser
	var stdout, stderr io.Reader

	stdin, err = session.StdinPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	stdout, err = session.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	stderr, err = session.StderrPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	wr := make(chan []byte, 10)

	go func() {
		for {
			select {
			case d := <-wr:
				_, err := stdin.Write(d)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for {
			if tkn := scanner.Scan(); tkn {
				rcv := scanner.Bytes()
				raw := make([]byte, len(rcv))
				copy(raw, rcv)
				fmt.Println(string(raw))
			} else {
				if scanner.Err() != nil {
					fmt.Println(scanner.Err())
				} else {
					fmt.Println("io.EOF")
				}
				return
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// execute cmd

	shellErr := session.Shell()
	failIfErr(shellErr, "Shell opening failed in "+c.Ip.String())

	for {
		fmt.Println("$")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text := scanner.Text()

		wr <- []byte(text + "\n")
	}

}

// e.g. output, err := remoteRun("root", "MY_IP", "PRIVATE_KEY", "ls")
func remoteRun(c *serverConfig, cmd string) (string, error) {
	// privateKey could be read from a file, or retrieved from another storage
	// source, such as the Secret Service / GNOME Keyring
	key, err := ssh.ParsePrivateKey(readFromFile(c.Key))
	if err != nil {
		return "", err
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
		return "", err
	}
	// Create a session. It is one session per command.
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var b bytes.Buffer  // import "bytes"
	session.Stdout = &b // get output
	// you can also pass what gets input to the stdin, allowing you to pipe
	// content from client to server
	//      session.Stdin = bytes.NewBufferString("My input")

	// Finally, run the command
	err = session.Run(cmd)
	return b.String(), err
}
