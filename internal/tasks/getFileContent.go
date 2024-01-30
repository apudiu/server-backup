package tasks

import (
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"strings"
)

// GetFileContent gets the file content from remote
func GetFileContent(c *ssh.Client, filePath string) (contents []byte, err error) {
	cmd := []string{
		"cat",
		filePath,
	}

	// create task for execution
	t := New(strings.Join(cmd, " "))
	contents, err = t.Execute(c)
	if err != nil {
		err = util.ErrWithPrefix("Error getting file content for "+c.RemoteAddr().String()+":"+filePath, err)
	}

	return
}
