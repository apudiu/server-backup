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

	conn, connErr := server.ConnectToServer(s)
	util.FailIfErr(connErr, "Connection establishment with server "+s.Ip.String()+" failed")
	defer conn.Close()

	// task
	p := s.Projects[0]

	// prepare paths
	sourcePath := p.SourcePath(s)
	sourceZipPath := sourcePath + ".zip" // dest path in remote
	destZipPath := p.DestPath(s) + util.DS + filepath.Base(sourceZipPath)
	projectLogFilePath := p.LogFilePath(s)

	// zip the dir
	_, err := tasks.ZipDirectory(conn, sourcePath, sourceZipPath, p.ExcludePaths, projectLogFilePath)
	util.FailIfErr(err, "Task failed...")

	// copy zip from server to local disk & log result
	copyLogger := logger.New()
	copyLogger.AddHeader("Copying " + sourceZipPath + " to " + destZipPath)

	_, err = server.GetFileFromServer(conn, sourceZipPath, destZipPath)
	if err != nil {
		copyLogger.AddHeader("Zip copy err: " + sourceZipPath + " to " + destZipPath + ". " + err.Error())
	} else {
		copyLogger.AddHeader("Copy Done: " + sourceZipPath + " to " + destZipPath)
	}

	err = copyLogger.WriteToFile(projectLogFilePath)
	util.FailIfErr(err, "Zip copy log failed to write in file")

	//fmt.Printf("Task %+v\n", tsk)
}
