package main

import (
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/logger"
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
	destZipPath := p.DestPath(s) + util.DS + filepath.Base(sourceZipPath)

	//ext, extErr := server.RemoteIsPathExist(conn, sourcePath)
	//fmt.Println("EE", ext, extErr)

	projectLogFilePath := p.LogFilePath(s)
	_, err := tasks.ZipDirectory(conn, sourcePath, sourceZipPath, p.ExcludePaths, projectLogFilePath)
	util.FailIfErr(err, "Task failed...")

	// copy zip from server
	copyLogger := logger.New()
	copyLogger.AddHeader([]byte("Copying " + sourceZipPath + " to " + destZipPath))

	_, err = server.GetFileFromServer(conn, sourceZipPath, destZipPath)
	if err != nil {
		copyLogger.AddHeader([]byte("Zip copy err: " + sourceZipPath + " to " + destZipPath + ". " + err.Error()))
	}

	err = copyLogger.WriteToFile(projectLogFilePath)
	util.FailIfErr(err, "Zip copy log failed to write in file")

	//fmt.Printf("Task %+v\n", tsk)
}
