package tasks

import (
	"golang.org/x/crypto/ssh"
	"strings"
)

// DeletePath deletes remote path
func DeletePath(c *ssh.Client, path string) (result []byte, err error) {
	cmd := []string{
		"rm -rf",
		path,
	}

	// create task for execution
	t := New(strings.Join(cmd, " "))
	result, err = t.Execute(c)
	return
}
