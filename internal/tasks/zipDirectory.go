package tasks

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"strings"
)

func ZipDirectory(c *ssh.Client, sourceDir, destZipPath string, excludeList []string) (t *Task, err error) {
	zipOptions := "-r9"
	zipOptions += formatExclude(excludeList)

	cmd := []string{
		// go to parent dir of the dir need to be zipped
		"cd",
		sourceDir + config.DS + "..",
		"&&",

		// zip the target dir
		"zip",
		zipOptions,
		destZipPath,
		filepath.Base(sourceDir),
	}

	t = New([]string{strings.Join(cmd, " ")})

	err = t.Execute(c)
	if err != nil {
		err = util.ErrWithPrefix("ZipDirectory task error for "+c.RemoteAddr().String(), err)
		return
	}

	fmt.Printf("t: %+v\n", t)

	return
}

func formatExclude(l []string) string {
	if l == nil {
		return ""
	}

	r := ` -x `
	for _, p := range l {
		r += `"` + p + `" `
	}
	return r
}
