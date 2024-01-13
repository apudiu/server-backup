package main

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func main() {
	generateEmptyConfigFile()

	//c := config{}
	//c.Parse()
	////fmt.Printf("%#v", c)
	//
	//// login to servers
	//for _, server := range c.Servers {
	//
	//}
}

func connectToServer(c *serverConfig) {
	serverKey := readFromFile(c.Key)

	signer, err := ssh.ParsePrivateKey(serverKey)
	if err != nil {
	todo:
	https: //medium.com/@marcus.murray/go-ssh-client-shell-session-c4d40daa46cd#:~:text=Go%20provides%20a%20package%20that,ClientConfig%20to%20be%20set%20up.
		if errors.Is(err, *ssh.PassphraseMissingError) {

		}

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
}
