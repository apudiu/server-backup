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
	srcBaseDir := filepath.Base(sourceDir)

	zipOptions := "-r9"
	exlcludeOptions := formatExclude(excludeList, srcBaseDir)

	cmd := []string{
		// go to parent dir of the dir need to be zipped
		"cd",
		sourceDir + config.DS + "..",
		"&&",

		// zip the target dir
		"zip",
		zipOptions,
		destZipPath,
		srcBaseDir,
		exlcludeOptions,
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

func formatExclude(l []string, prefixPath string) string {
	if l == nil {
		return ""
	}

	lLen := len(l) - 1

	r := `-x `
	for i, p := range l {
		r += fmt.Sprintf(`"%s%s%s"`, prefixPath, config.DS, p)

		// add space after each except last one
		if i < lLen {
			r += " "
		}
	}
	return r
}
