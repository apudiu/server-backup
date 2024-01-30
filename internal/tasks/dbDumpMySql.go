package tasks

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/logger"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

func DbDumpMySql(
	c *ssh.Client,
	sc *config.ServerConfig,
	pc *config.ProjectConfig,
	l *logger.Logger,
) (t *Task, err error) {
	srcDir := pc.SourcePath(sc)

	cmdOptions := fmt.Sprintf(
		`-e -v -h"%v" -u"%v" -p"%v" --add-drop-table`,
		pc.DbInfo.Host, pc.DbInfo.User, pc.DbInfo.Pass,
	)

	// like: 2024-12-37_15-16-17_db_name.sql
	dbDumpFileName := time.Now().Format(time.DateTime)
	dbDumpFileName = strings.ReplaceAll(dbDumpFileName, " ", "_")
	dbDumpFileName = strings.ReplaceAll(dbDumpFileName, ":", "-")
	dbDumpFileName += "_" + pc.DbInfo.Name + ".sql"

	cmd := []string{
		// go to parent dir of the dir need to be zipped
		"cd",
		srcDir + util.DS + "..",
		"&&",

		// zip the target dir
		"mysqldump",
		cmdOptions,
		pc.DbInfo.Name,
		">",
		dbDumpFileName,
	}

	fmt.Println("cmd: ", strings.Join(cmd, " "))

	// create task for execution
	t = New(strings.Join(cmd, " "))
	start, wait, closeFn, err := t.ExecuteLive(c)
	if err != nil {
		err = util.ErrWithPrefix("DB dump task error for "+c.RemoteAddr().String(), err)
		return
	}
	defer closeFn()

	// read output in realtime
	l.AddHeader("Dumping " + pc.DbInfo.Name + " from " + sc.Ip.String() + sc.ProjectRoot + "/" + pc.Path)

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
