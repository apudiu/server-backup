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

// ExecCmd executes cmd on the connection, Generally used with go routines.
// Caller need to wait the returned session to finish the task
// after reading all outputs from stdOut & stdErr.
// if call wait before reading all outputs then some output might get cut off
func ExecCmd(conn *ssh.Client, cmd string) (result []byte, err error) {

	var buf io.Reader
	//reader := bytes.NewReader(buf)

	start, wait, closeFn, err := MakeExecCmd(conn, cmd, &buf)
	if err != nil {
		return
	}
	defer func() {
		err = closeFn()
	}()

	err = start()

	err = wait()

	return
}

func MakeExecCmd(
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

	//if stdErr != nil {
	//	*stdErr, err = session.StderrPipe()
	//	if err != nil {
	//		return
	//	}
	//}

	wait = session.Wait
	start = func() error {
		return session.Start(cmd)
	}
	closeFn = session.Close
	return
}

//func RemoteIsPathExist(c *ssh.Client, p string) (bool, error) {
//	cmd := "ls " + p
//
//	wait, err := ExecCmd(c, cmd)
//	if err != nil {
//		return false, err
//	}
//
//	err = wait()
//	if err != nil {
//		return false, err
//	}
//
//	return true, nil
//}

func GerFileFromServer(c *ssh.Client, sourcePath, destPath string) (isSuccess bool, err error) {
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
	df, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		err = util.ErrWithPrefix("Dest file creation error on", err)
		return
	}
	defer func(df *os.File) {
		e := df.Close()
		if e != nil {
			err = util.ErrWithPrefix("Local file closing err", e)
		}
	}(df)

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
	//todo: test zip from root dir
	return true, nil
}
