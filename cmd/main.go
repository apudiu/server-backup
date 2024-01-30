package main

import (
	"fmt"
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
	envFilePath := s.ProjectRoot + util.DS + p.Path + util.DS + p.EnvFileInfo.Path

	// logger
	l := logger.New()

	// zip the dir
	_, err := tasks.ZipDirectory(conn, sourcePath, sourceZipPath, p.ExcludePaths, projectLogFilePath)
	util.FailIfErr(err, "Task failed...")

	// copy zip from server to local disk & log result
	l.AddHeader("Copying " + sourceZipPath + " to " + destZipPath)

	_, err = server.GetFileFromServer(conn, sourceZipPath, destZipPath)
	if err != nil {
		l.AddHeader("Zip copy err: " + sourceZipPath + " to " + destZipPath + ". " + err.Error())
	} else {
		l.AddHeader("Copy Done: " + sourceZipPath + " to " + destZipPath)
	}

	// do db backup

	// parse server's env
	fc, err := tasks.GetFileContent(conn, envFilePath)
	if err != nil {
		fmt.Println("Env file content err", err.Error())
		return
	}

	err = p.ParseDbInfo(fc, '\n')
	if err != nil {
		fmt.Println("Env parse err")
	}

	_, err = tasks.DbDumpMySql(conn, s, &p, l)
	if err != nil {
		fmt.Println("err........", err.Error())
	}
	//util.FailIfErr(err)

	//todo: download db backup
	// write all logs to file
	err = l.WriteToFile(projectLogFilePath)
	util.FailIfErr(err, "Failed to write in log file")
}
