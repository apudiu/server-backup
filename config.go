package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DS              = string(os.PathSeparator)
	configDir       = "." + DS + "config"
	serverConfigFle = configDir + DS + "servers.yml"
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

type serverConfig struct {
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

type config struct {
	Servers []serverConfig `yaml:"servers"`
}

func (c *config) Parse() {
	if exists, _ := isPathExist(serverConfigFle); !exists {
		log.Fatalln("Config file unavailable at " + serverConfigFle)
	}

	sb, err := os.ReadFile(serverConfigFle)
	failIfErr(err)

	unmarshalErr := yaml.Unmarshal(sb, c)
	failIfErr(unmarshalErr)

	// load server projects
	for _, server := range c.Servers {
		server.Projects = make([]projectPath, 1)
		projects := make([]projectPath, 0)
		configFileDir := configDir + DS + server.Ip.String() + DS

		for _, sourceDir := range server.BackupSources {
			projectConfigFile := configFileDir + sourceDir + ".yml"

			if exists, accessErr := isPathExist(projectConfigFile); !exists {
				if accessErr != nil {
					log.Println(accessErr)
				}
				log.Println("Project config file unavailable at " + projectConfigFile + " SKIPPING!")
				continue
			}

			pb, pbErr := os.ReadFile(projectConfigFile)
			if pbErr != nil {
				log.Println(pbErr)
				log.Println("Project config file parse err " + projectConfigFile + " SKIPPING!")
				continue
			}

			pc := projectPath{}

			unmarshalProjectErr := yaml.Unmarshal(pb, &pc)
			if unmarshalProjectErr != nil {
				log.Println(unmarshalProjectErr)
				log.Println("Project config file invalid " + projectConfigFile + " SKIPPING!")
				continue
			}

			//todo: parse projects into server
			projects = append(projects, pc)
			server.Projects = append(server.Projects, pc)
		}

		server.Projects = append(server.Projects, projects...)
	}
}

func generateEmptyConfigFile() {

	if configFileExists, _ := isPathExist(serverConfigFle); configFileExists {
		input, err := readUserInput("Server config file exist, overwrite? [y/N]")
		failIfErr(err, "Failed to read your input")

		input = strings.ToLower(input)

		if input == "n" {
			fmt.Println("Cancelled!")
			return
		}
	} else {
		// config dir exist
		if configDirExist, _ := isPathExist(configDir); !configDirExist {
			err := os.Mkdir(configDir, 0755)
			failIfErr(err, "Config dir creation err")
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
	servers := config{
		Servers: []serverConfig{
			{
				Key:            makeAbsoluteFilePath("home", "user", "serverKey.pem"),
				Ip:             net.IP{192, 168, 0, 100},
				Port:           22,
				User:           "privilegedUserWhoCanDoYourTasks",
				Password:       "123456",
				ProjectRoot:    makeAbsoluteFilePath("var", "www", "php80"),
				BackupDestPath: "",
				BackupSources:  []string{},
			},
		},
	}

	// add projects to server
	for _, project := range projects {
		servers.Servers[0].BackupSources = append(servers.Servers[0].BackupSources, project.Path)
	}

	// write servers config
	serversYmlData, err := yaml.Marshal(&servers)
	failIfErr(err)
	writeToFile(serverConfigFle, serversYmlData, 0600)

	// write projects config
	for _, project := range projects {
		projectYmlData, errS := yaml.Marshal(&project)
		failIfErr(errS)

		// create dir if not exist
		projectDir := configDir + DS + servers.Servers[0].Ip.String()
		if projectDirExist, _ := isPathExist(projectDir); !projectDirExist {
			err2 := os.MkdirAll(projectDir, 0755)
			failIfErr(err2, "Project config dir creation err: "+projectDir)
		}
		projectConfigFile := projectDir + DS + project.Path + ".yml"
		writeToFile(projectConfigFile, projectYmlData, 0600)
	}
}

func (c *projectPath) parseDbInfo() {
	if c.EnvFileInfo.Path == "" {
		return
	}

	envEntries := parseEnv(c.EnvFileInfo.Path)

	host := envEntries[c.EnvFileInfo.DbHostKeyName]
	if host != "" {
		c.DbInfo.Host = net.ParseIP(host)
	}

	portStr := envEntries[c.EnvFileInfo.DbPortKeyName]
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		failIfErr(err, "Failed to parse DB Port for "+host)
		c.DbInfo.Port = port
	}

	c.DbInfo.User = envEntries[c.EnvFileInfo.DbUserKeyName]
	c.DbInfo.Pass = envEntries[c.EnvFileInfo.DbPassKeyName]
	c.DbInfo.Name = envEntries[c.EnvFileInfo.DbNameKeyName]
}
