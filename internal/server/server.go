package server

import (
	"context"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/util"
	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
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

// ExecCmd executes cmd and returns the response
func ExecCmd(conn *ssh.Client, cmd string) (result []byte, err error) {
	session, err := conn.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	result, err = session.CombinedOutput(cmd)
	return
}

// ExecCmdLive executes cmd and pipes stdOut & stdErr in @stdOutErr for live reading.
// Caller need to call @start to start execution, call @wait to let the command finish
// and before calling wait output stream can be read & finally call @closeFn (defer)
func ExecCmdLive(
	conn *ssh.Client, cmd string, stdOutErr *io.Reader,
) (start, wait, closeFn func() error, err error) {
	session, err := conn.NewSession()
	if err != nil {
		return
	}

	session.Stdout = session.Stderr

	// process CMD
	if stdOutErr != nil {
		*stdOutErr, err = session.StdoutPipe()
		if err != nil {
			return
		}
	}

	wait = session.Wait
	start = func() error {
		return session.Start(cmd)
	}
	closeFn = session.Close
	return
}

// RemoteIsPathExist checks if remote path exists
func RemoteIsPathExist(c *ssh.Client, p string) (bool, error) {
	cmd := "ls " + p

	_, err := ExecCmd(c, cmd)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetFileFromServer(c *ssh.Client, sourcePath, destPath string) (success bool, err error) {
	// check if remote file exists
	//exist, err := RemoteIsPathExist(c, sourcePath)
	//if err != nil || !exist {
	//	err = util.ErrWithPrefix("Remote file missing", err)
	//	return
	//}
	//
	client, err := scp.NewClientBySSH(c)
	if err != nil {
		err = util.ErrWithPrefix("Failed to get server session", err)
		return
	}
	defer client.Close()

	// download the file

	// open local file to write to it
	df, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		err = util.ErrWithPrefix("Dest file creation error on", err)
		return
	}
	defer df.Close()

	err = client.CopyFromRemote(context.Background(), df, sourcePath)
	if err != nil {
		err = util.ErrWithPrefix("File transfer failed for "+sourcePath, err)
		return
	}

	// verify file
	isExist, err := util.IsPathExist(destPath)
	if err != nil || !isExist {
		err = util.ErrWithPrefix("Result file might not be usable ", err)
		return
	}

	return true, nil
}
