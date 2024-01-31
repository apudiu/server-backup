package config

import (
	"errors"
	"fmt"
	"github.com/apudiu/server-backup/internal/util"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type projectDbInfo struct {
	Host net.IP `yaml:"hostIp"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
	Name string `yaml:"name"`
}

type projectEnvFileInfo struct {
	Path          string `yaml:"path"`
	DbHostKeyName string `yaml:"dbHostKeyName"`
	DbPortKeyName string `yaml:"dbPortKeyName"`
	DbUserKeyName string `yaml:"dbUserKeyName"`
	DbPassKeyName string `yaml:"dbPassKeyName"`
	DbNameKeyName string `yaml:"dbNameKeyName"`
}

type ProjectConfig struct {
	Path         string             `yaml:"path"`
	ExcludePaths []string           `yaml:"excludePaths"`
	ZipFileName  string             `yaml:"zipFileName"`
	EnvFileInfo  projectEnvFileInfo `yaml:"envFileInfo"`

	// when env file is not available, provide DB credentials
	DbInfo projectDbInfo `yaml:"dbInfo"`
}

type ServerConfig struct {
	Key            string   `yaml:"privateKeyPath"`
	Ip             net.IP   `yaml:"ip"`
	Port           int      `yaml:"port"`
	User           string   `yaml:"user"`
	Password       string   `yaml:"password"`
	ProjectRoot    string   `yaml:"projectRoot"`
	BackupSources  []string `yaml:"backupSources"`
	BackupDestPath string   `yaml:"backupDestPath"`
	Projects       []ProjectConfig
}

type Config struct {
	Servers []ServerConfig `yaml:"servers"`
}

// Parse parses configs for all servers and projects under them
func (c *Config) Parse() {
	if exists, _ := util.IsPathExist(util.ServerConfigFle); !exists {
		log.Fatalln("Config file unavailable at " + util.ServerConfigFle)
	}

	sb, err := os.ReadFile(util.ServerConfigFle)
	util.FailIfErr(err)

	unmarshalErr := yaml.Unmarshal(sb, c)
	util.FailIfErr(unmarshalErr)

	// load server projects
	for srvIdx := range c.Servers {
		server := &c.Servers[srvIdx]
		configFileDir := util.ConfigDir + util.DS + server.Ip.String() + util.DS

		for _, sourceDir := range server.BackupSources {
			projectConfigFile := configFileDir + sourceDir + ".yml"

			if exists, accessErr := util.IsPathExist(projectConfigFile); !exists {
				if accessErr != nil {
					log.Println(accessErr)
				}
				log.Println("Project Config file unavailable at " + projectConfigFile + " SKIPPING!")
				continue
			}

			pb, pbErr := os.ReadFile(projectConfigFile)
			if pbErr != nil {
				log.Println(pbErr)
				log.Println("Project Config file parse err " + projectConfigFile + " SKIPPING!")
				continue
			}

			pc := ProjectConfig{}

			unmarshalProjectErr := yaml.Unmarshal(pb, &pc)
			if unmarshalProjectErr != nil {
				log.Println(unmarshalProjectErr)
				log.Println("Project Config file invalid " + projectConfigFile + " SKIPPING!")
				continue
			}

			server.Projects = append(server.Projects, pc)
		}
	}
}

// ParseDbInfo tries to parse DB info form provided .env file
func (pc *ProjectConfig) ParseDbInfo(envContent []byte, envEol byte) error {
	if pc.EnvFileInfo.Path == "" {
		return errors.New("env file is not specified")
	}

	if envContent == nil || len(envContent) == 0 {
		return errors.New("envContent missing")
	}

	envEntries, ok := util.ParseEnvFromContent(envContent, envEol)
	if !ok {
		return errors.New("failed to parse env")
	}

	host := envEntries[pc.EnvFileInfo.DbHostKeyName]
	if host != "" {
		pc.DbInfo.Host = net.ParseIP(host)
	}

	portStr := envEntries[pc.EnvFileInfo.DbPortKeyName]
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		util.FailIfErr(err, "Failed to parse DB Port for "+host)
		pc.DbInfo.Port = port
	}

	pc.DbInfo.User = envEntries[pc.EnvFileInfo.DbUserKeyName]
	pc.DbInfo.Pass = envEntries[pc.EnvFileInfo.DbPassKeyName]
	pc.DbInfo.Name = envEntries[pc.EnvFileInfo.DbNameKeyName]

	return nil
}

// SourcePath returns project remote absolute path
func (pc *ProjectConfig) SourcePath(sc *ServerConfig) string {
	return sc.ProjectRoot + util.DS + pc.Path // server/path/project/path
}

// DestPath returns project local absolute path
func (pc *ProjectConfig) DestPath(sc *ServerConfig) string {
	p := sc.BackupDestPath

	// if dest path not specified use default
	if p == "" {
		p = util.BackupDir + util.DS + sc.Ip.String()
	}

	// if dest path doesn't contain trailing slash, add that
	lc := p[len(p)-1:]
	if lc != "/" || lc != util.DS {
		p += util.DS
	}

	return p + pc.Path + util.DS + time.Now().Format(time.DateOnly) // source/path/project/path
}

// LogFilePath returns local log file path
func (pc *ProjectConfig) LogFilePath(sc *ServerConfig) string {
	return pc.DestPath(sc) + util.DS + time.Now().Format(time.DateOnly) + ".log"
}

// DbDumpFilePath returns db dump file's absolute path
// like: /path/to/server/path/to/project/2024-12-17_120925_db_name.sql.gz
func (pc *ProjectConfig) DbDumpFilePath(sc *ServerConfig) string {
	f := time.Now().Format(time.DateOnly)
	f += "_" + pc.DbInfo.Name + ".sql.gz"

	return filepath.Dir(pc.SourcePath(sc)) + util.DS + f
}

// GenerateEmptyConfigFile generates sample config files
func GenerateEmptyConfigFile() {

	if configFileExists, _ := util.IsPathExist(util.ServerConfigFle); configFileExists {
		input, err := util.ReadUserInput("Server Config file exist, overwrite? [y/N]")
		util.FailIfErr(err, "Failed to read your input")

		input = strings.ToLower(input)

		if input == "n" {
			fmt.Println("Cancelled!")
			return
		}
	} else {
		// Config dir exist
		if configDirExist, _ := util.IsPathExist(util.ConfigDir); !configDirExist {
			err := os.Mkdir(util.ConfigDir, 0755)
			util.FailIfErr(err, "Config dir creation err")
		}
	}

	projects := []ProjectConfig{
		{
			Path: "order-online",
			ExcludePaths: []string{
				filepath.Join("api", "vendor"),
				filepath.Join("api", "storage", "app", "*"),
				filepath.Join("api", "storage", "framework", "*"),
				filepath.Join("api", "storage", "logs", "*"),
				filepath.Join("api", ".rsyncIgnore"),
				filepath.Join("www", "vendor", "*"),
			},
			ZipFileName: "",
			EnvFileInfo: projectEnvFileInfo{
				Path:          filepath.Join("api", ".env"),
				DbHostKeyName: "DB_HOST",
				DbPortKeyName: "DB_PORT",
				DbUserKeyName: "DB_USERNAME",
				DbPassKeyName: "DB_PASSWORD",
				DbNameKeyName: "DB_DATABASE",
			},
			DbInfo: projectDbInfo{
				Host: nil,
				User: "",
				Pass: "",
				Name: "",
			},
		},
		{
			Path: "buy-sell",
			ExcludePaths: []string{
				filepath.Join("api", "vendor"),
				filepath.Join("api", "storage", "app", "*"),
				filepath.Join("api", "storage", "framework", "*"),
				filepath.Join("api", "storage", "logs", "*"),
				filepath.Join("api", ".rsyncIgnore"),
				filepath.Join("www", "vendor", "*"),
			},
			ZipFileName: "",
			EnvFileInfo: projectEnvFileInfo{
				Path:          filepath.Join("api", ".env"),
				DbHostKeyName: "DB_HOST",
				DbPortKeyName: "DB_PORT",
				DbUserKeyName: "DB_USERNAME",
				DbPassKeyName: "DB_PASSWORD",
				DbNameKeyName: "DB_DATABASE",
			},
			DbInfo: projectDbInfo{
				Host: nil,
				Port: 0,
				User: "",
				Pass: "",
				Name: "",
			},
		},
	}

	// generated project info
	servers := Config{
		Servers: []ServerConfig{
			{
				Key:            util.MakeAbsoluteFilePath("home", "user", "serverKey.pem"),
				Ip:             net.IP{192, 168, 0, 100},
				Port:           22,
				User:           "privilegedUserWhoCanDoYourTasks",
				Password:       "123456",
				ProjectRoot:    util.MakeAbsoluteFilePath("var", "www", "php80"),
				BackupDestPath: "",
				BackupSources:  []string{},
			},
		},
	}

	// add projects to server
	for _, project := range projects {
		servers.Servers[0].BackupSources = append(servers.Servers[0].BackupSources, project.Path)
	}

	// write servers Config
	serversYmlData, err := yaml.Marshal(&servers)
	util.FailIfErr(err)
	util.WriteToFile(util.ServerConfigFle, serversYmlData, 0600)

	// write projects Config
	for _, project := range projects {
		projectYmlData, errS := yaml.Marshal(&project)
		util.FailIfErr(errS)

		// create dir if not exist
		projectDir := util.ConfigDir + util.DS + servers.Servers[0].Ip.String()
		if projectDirExist, _ := util.IsPathExist(projectDir); !projectDirExist {
			err2 := os.MkdirAll(projectDir, 0755)
			util.FailIfErr(err2, "Project Config dir creation err: "+projectDir)
		}
		projectConfigFile := projectDir + util.DS + project.Path + ".yml"
		util.WriteToFile(projectConfigFile, projectYmlData, 0600)
	}
}
