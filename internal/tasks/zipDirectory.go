package tasks

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/logger"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"strings"
)

func ZipDirectory(
	c *ssh.Client,
	sourceDir, destZipPath string,
	excludeList []string,
	logFilePath string,
) (t *Task, err error) {
	srcBaseDir := filepath.Base(sourceDir)

	zipOptions := "-r9"
	excludeOptions := formatExclude(excludeList, srcBaseDir)

	cmd := []string{
		// go to parent dir of the dir need to be zipped
		"cd",
		sourceDir + util.DS + "..",
		"&&",

		// zip the target dir
		"zip",
		zipOptions,
		destZipPath,
		srcBaseDir,
		excludeOptions,
	}

	// create task for execution
	t = New(strings.Join(cmd, " "))
	start, wait, closeFn, err := t.ExecuteLive(c)
	if err != nil {
		err = util.ErrWithPrefix("ZipDirectory task error for "+c.RemoteAddr().String(), err)
		return
	}
	defer closeFn()

	// read output in realtime
	l := logger.New()
	l.AddHeader([]byte("Zipping"))

	ch := make(chan struct{})
	go func() {
		l.ReadStream(&t.StdOutErr)
		ch <- struct{}{}
	}()

	// wait to copy all output
	err = start()
	<-ch

	// wait to finish the task
	err = wait()

	// put all outputs to the file
	err = l.WriteToFile(logFilePath)
	return
}

func formatExclude(l []string, prefixPath string) string {
	if l == nil {
		return ""
	}

	lLen := len(l) - 1

	r := `-x `
	for i, p := range l {
		r += fmt.Sprintf(`"%s%s%s"`, prefixPath, util.DS, p)

		// add space after each except last one
		if i < lLen {
			r += " "
		}
	}
	return r
}
