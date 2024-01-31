package tasks

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/logger"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"strings"
)

func DbDumpMySql(
	c *ssh.Client,
	sc *config.ServerConfig,
	pc *config.ProjectConfig,
	l *logger.Logger,
	dumpFilePath string,
) (t *Task, err error) {
	srcDir := pc.SourcePath(sc)

	cmdOptions := fmt.Sprintf(
		`-e -h"%v" -u"%v" -p"%v" --add-drop-table`,
		pc.DbInfo.Host, pc.DbInfo.User, pc.DbInfo.Pass,
	)

	cmd := []string{
		// go to parent dir of the dir need to be zipped
		"cd",
		srcDir + util.DS + "..",
		"&&",

		// zip the target dir
		"mysqldump",
		cmdOptions,
		pc.DbInfo.Name,
		"|",
		"gzip -9",
		">",
		dumpFilePath,
	}

	// create task for execution
	t = New(strings.Join(cmd, " "))
	start, wait, closeFn, err := t.ExecuteLive(c)
	if err != nil {
		err = util.ErrWithPrefix("DB dump task error for "+c.RemoteAddr().String(), err)
		return
	}
	defer closeFn()

	// read output in realtime

	l.AddHeader(
		fmt.Sprintf("Dumping %s from %s:%s", pc.DbInfo.Name, sc.Ip.String(), sc.ProjectRoot+util.DS+pc.Path),
	)

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

	return
}
