package main

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/server"
	"github.com/apudiu/server-backup/internal/tasks"
	"github.com/apudiu/server-backup/internal/util"
	"path/filepath"
)

func main() {
	//generateEmptyConfigFile()

	c := config.Config{}
	c.Parse()
	//fmt.Printf("%+v \n", c)

	for _, srv := range c.Servers {
		doWork(&srv)
	}

	//realtimeRead()

}

func doWork(s *config.ServerConfig) {
	//stdOut, stdErr, err := remoteRun(s, "ls /var/www/php80")
	//fmt.Println("stdOut", stdOut)
	//fmt.Println("stdErr", stdErr)
	//fmt.Println("err", err)

	//l := logger.Logger{}
	//l.AddLn([]byte("Alhum-du-lillah"))
	//l.Add([]byte("Subhan Allah"))
	//
	//er := l.WriteToFile("./logs.log")
	//if er != nil {
	//	fmt.Println("Log err", er.Error())
	//}

	conn, connErr := server.ConnectToServer(s)
	util.FailIfErr(connErr, "Connection establishment with server "+s.Ip.String()+" failed")
	defer conn.Close()

	// task
	p := s.Projects[0]

	//sourcePath := s.ProjectRoot + config.DS + p.Path // source path
	sourcePath := p.SourcePath(s)
	sourceZipPath := sourcePath + ".zip" // dest path in remote
	destZipPath := filepath.Dir(p.DestPath(s)) + util.DS + filepath.Base(sourceZipPath)

	//ext, extErr := server.RemoteIsPathExist(conn, sourcePath)
	//fmt.Println("EE", ext, extErr)

	zipLogPath := p.LogFilePath(s)
	_, err := tasks.ZipDirectory(conn, sourcePath, sourceZipPath, p.ExcludePaths, zipLogPath)
	util.FailIfErr(err, "Task failed...")

	fmt.Println(sourceZipPath, destZipPath)

	//fmt.Printf("Task %+v\n", tsk)
}
