package main

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/config"
	"github.com/apudiu/server-backup/internal/logger"
	"github.com/apudiu/server-backup/internal/remotebackup"
	"github.com/apudiu/server-backup/internal/server"
	"github.com/apudiu/server-backup/internal/tasks"
	"github.com/apudiu/server-backup/internal/util"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"sync"
)

func main() {
	remotebackup.ListBucket()
	return

	runLog := logger.New()
	runLog.ToggleStdOut(true)
	runLog.AddHeader(util.ServerLogf("Starting backup"))

	//generateEmptyConfigFile()

	c := config.Config{}
	c.Parse()
	//fmt.Printf("%+v \n", c)

	wg := sync.WaitGroup{}
	wg.Add(len(c.Servers))

	for si := range c.Servers {
		go func(s *config.ServerConfig) {
			runLog.AddHeader(util.ServerLogf("Processing server: " + s.Ip.String()))
			processServer(s, runLog)
			runLog.AddHeader(util.ServerLogf("Processed server: " + s.Ip.String()))
			wg.Done()
		}(&c.Servers[si])
	}

	wg.Wait()
	runLog.AddHeader("Backup ended")

	runLogFilePath := util.BackupDir + util.DS + "run.log"

	err := runLog.WriteToFile(runLogFilePath)
	if err != nil {
		log.Println("Failed to write to run.log", err.Error())
		os.Exit(1)
	}

}

func processServer(s *config.ServerConfig, runLogger *logger.Logger) {

	conn, connErr := server.ConnectToServer(s)
	if connErr != nil {
		runLogger.AddHeader(
			util.ServerLogLn("Connection establishment with server", s.Ip.String(), "failed.", connErr.Error()),
		)
		return
	}
	defer conn.Close()

	wg := sync.WaitGroup{}
	wg.Add(len(s.Projects))

	for pi := range s.Projects {
		go func(p *config.ProjectConfig) {
			projOnSrvPathStr := fmt.Sprintf("%s:%s", s.Ip.String(), p.SourcePath(s))
			runLogger.AddHeader(util.ProjectLogf("Processing project: %s", projOnSrvPathStr))

			er := processProject(conn, s, p)
			if er != nil {
				runLogger.AddHeader(
					util.ProjectFailLogLn("Processing project failed", projOnSrvPathStr, er.Error()),
				)
			} else {
				runLogger.AddHeader(util.ProjectLogf("Processed project: " + projOnSrvPathStr))
			}

			wg.Done()
		}(&s.Projects[pi])
	}

	wg.Wait()
}

func processProject(
	conn *ssh.Client,
	sc *config.ServerConfig,
	pc *config.ProjectConfig,
) error {
	// logger
	l := logger.New()

	// prepare paths
	localPath := pc.DestPath(sc)
	err := util.CreatePath(localPath, 0755, false)
	if err != nil {
		l.AddHeader("Failed to create local path. " + err.Error())
		logErr := l.WriteToFile(pc.LogFilePath(sc))
		if logErr != nil {
			log.Println("Failed to write in log file", logErr.Error())
		}
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	// zip the dir
	go func() {
		zipAndCopyFiles(conn, sc, pc, l)
		wg.Done()
	}()

	// do db backup
	go func() {
		dumpDdAndCopy(conn, sc, pc, l)
		wg.Done()
	}()

	wg.Wait()

	// write all logs to file
	err = l.WriteToFile(pc.LogFilePath(sc))
	if err != nil {
		log.Println(err, "Failed to write in log file")
		return err
	}

	return nil
}

func zipAndCopyFiles(
	conn *ssh.Client,
	s *config.ServerConfig,
	p *config.ProjectConfig,
	l *logger.Logger,
) {
	remotePath := p.SourcePath(s)
	remoteZipPath, localZipPath := p.ZipFilePath(s)

	_, err := tasks.ZipDirectory(conn, remotePath, remoteZipPath, p.ExcludePaths, l)
	if err != nil {
		l.AddHeader(fmt.Sprintf("Ziping failed for %s", remotePath))
		return
	}

	// copy zip from server to local disk & log result
	l.AddHeader(fmt.Sprintf("Copying: %s --> %s", remoteZipPath, localZipPath))

	_, err = server.GetFileFromServer(conn, remoteZipPath, localZipPath)
	if err != nil {
		l.AddHeader(fmt.Sprintf("Copy err: %s --> %s. %s", remoteZipPath, localZipPath, err.Error()))
	} else {
		l.AddHeader(fmt.Sprintf("Copy Done: %s --> %s", remoteZipPath, localZipPath))
	}
}

func dumpDdAndCopy(
	conn *ssh.Client,
	s *config.ServerConfig,
	p *config.ProjectConfig,
	l *logger.Logger,
) {
	remoteEnvPath := s.ProjectRoot + util.DS + p.Path + util.DS + p.EnvFileInfo.Path
	envContent, err := tasks.GetFileContent(conn, remoteEnvPath)
	if err != nil {
		l.AddHeader("Error getting env file. " + err.Error())
		return
	}

	err = p.ParseDbInfo(envContent, '\n')
	if err != nil {
		l.AddHeader("Error parsing env content. " + err.Error())
		return
	}

	remoteDbDumpPath, localDbDumpPath := p.DbDumpFilePath(s)

	_, err = tasks.DbDumpMySql(conn, s, p, l, remoteDbDumpPath)
	if err != nil {
		l.AddHeader("DB dumping error. " + err.Error())
		return
	}

	// download db dump
	l.AddHeader("Copying " + remoteDbDumpPath + " to " + localDbDumpPath)

	_, err = server.GetFileFromServer(conn, remoteDbDumpPath, localDbDumpPath)
	if err != nil {
		l.AddHeader(fmt.Sprintf("DB dump copy err: %s --> %s. %s", remoteDbDumpPath, localDbDumpPath, err.Error()))
	} else {
		l.AddHeader(fmt.Sprintf("Copy done: %s --> %s", remoteDbDumpPath, localDbDumpPath))
	}
}
