package tasks

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"strings"
)

func ZipDirectory(c *ssh.Client, sourceDir, destZipPath string) (t *Task, err error) {
	cmd := []string{
		"zip -r",
		destZipPath,
		sourceDir,
	}

	t = New([]string{strings.Join(cmd, " ")})

	err = t.Execute(c)
	if err != nil {
		err = util.ErrWithPrefix("ZipDirectory task error for "+c.RemoteAddr().String(), err)
		return
	}

	fmt.Println("After zip")

	fmt.Printf("t: %+v\n", t)

	return
}
