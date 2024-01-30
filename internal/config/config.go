package config

import (
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

type projectPath struct {
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
	Projects       []projectPath
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

			pc := projectPath{}

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

// parseDbInfo tries to parse DB info form provided .env file
func (c *projectPath) parseDbInfo() {
	if c.EnvFileInfo.Path == "" {
		return
	}

	envEntries := util.ParseEnv(c.EnvFileInfo.Path)

	host := envEntries[c.EnvFileInfo.DbHostKeyName]
	if host != "" {
		c.DbInfo.Host = net.ParseIP(host)
	}

	portStr := envEntries[c.EnvFileInfo.DbPortKeyName]
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		util.FailIfErr(err, "Failed to parse DB Port for "+host)
		c.DbInfo.Port = port
	}

	c.DbInfo.User = envEntries[c.EnvFileInfo.DbUserKeyName]
	c.DbInfo.Pass = envEntries[c.EnvFileInfo.DbPassKeyName]
	c.DbInfo.Name = envEntries[c.EnvFileInfo.DbNameKeyName]
}

// SourcePath returns project remote absolute path
func (c *projectPath) SourcePath(s *ServerConfig) string {
	return s.ProjectRoot + util.DS + c.Path // server/path/project/path
}

// DestPath returns project local absolute path
func (c *projectPath) DestPath(s *ServerConfig) string {
	p := s.BackupDestPath

	// if dest path not specified use default
	if p == "" {
		p = util.BackupDir + util.DS + s.Ip.String()
	}

	// if dest path doesn't contain trailing slash, add that
	lc := p[len(p)-1:]
	if lc != "/" || lc != util.DS {
		p += util.DS
	}

	return p + c.Path // source/path/project/path
}

// LogFilePath returns local log file path
func (c *projectPath) LogFilePath(s *ServerConfig) string {
	return c.DestPath(s) + util.DS + time.Now().Format(time.DateOnly) + ".log"
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

	projects := []projectPath{
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
