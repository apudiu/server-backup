package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"os"
	"path"
	"strings"
)

var configFilePath = "./config.yml"

type projectDbInfo struct {
	Host net.IP `yaml:"hostIp"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
	Name string `yaml:"name"`
}

type projectPath struct {
	Path         string   `yaml:"path"`
	ExcludePaths []string `yaml:"excludePaths"`
	ZipFileName  string   `yaml:"zipFileName"`
	EnvFilePath  string   `yaml:"envFilePath"`

	// when env file is not available, provide DB credentials
	DbInfo projectDbInfo `yaml:"dbInfo"`
}

type serverConfig struct {
	Key            string        `yaml:"privateKeyPath"`
	Ip             net.IP        `yaml:"ip"`
	Port           int           `yaml:"port"`
	ProjectRoot    string        `yaml:"projectRoot"`
	BackupSources  []projectPath `yaml:"BackupSources"`
	BackupDestPath string        `yaml:"backupDestPath"`
}

type config struct {
	Servers []serverConfig `yaml:"servers"`
}

func (c *config) Parse() {
	if exists, _ := isPathExist(configFilePath); !exists {
		log.Fatalln("Config file unavailable at " + configFilePath)
	}

	fb, err := os.ReadFile(configFilePath)
	failIfErr(err)

	unmarshalErr := yaml.Unmarshal(fb, c)
	failIfErr(unmarshalErr)
}

func generateEmptyConfigFile() {

	if exists, _ := isPathExist(configFilePath); exists {
		input, err := readUserInput("Config file exist, overwrite? [y/N]")
		failIfErr(err, "Failed to read your input")

		input = strings.ToLower(input)

		if input == "n" {
			fmt.Println("Cancelled!")
			return
		}
	}

	// generated project info
	c := config{
		Servers: []serverConfig{
			{
				Key:            "/home/user/serverKey.pem",
				Ip:             net.IP{192, 168, 0, 100},
				Port:           21,
				ProjectRoot:    path.Join("var", "www", "php80"),
				BackupDestPath: "",
				BackupSources: []projectPath{
					{
						Path: "order-online",
						ExcludePaths: []string{
							path.Join("api", "vendor"),
							path.Join("api", "storage", "app", "*"),
							path.Join("api", "storage", "framework", "*"),
							path.Join("api", "storage", "logs", "*"),
							path.Join("api", ".rsyncIgnore"),
						},
						ZipFileName: "",
						EnvFilePath: path.Join("api", ".env"),
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
							path.Join("api", "vendor"),
							path.Join("api", "storage", "app", "*"),
							path.Join("api", "storage", "framework", "*"),
							path.Join("api", "storage", "logs", "*"),
							path.Join("api", ".rsyncIgnore"),
						},
						ZipFileName: "",
						EnvFilePath: path.Join("api", ".env"),
						DbInfo: projectDbInfo{
							Host: nil,
							User: "",
							Pass: "",
							Name: "",
						},
					},
				},
			},
		},
	}

	ymlData, err := yaml.Marshal(&c)
	failIfErr(err)

	writeToFile(configFilePath, ymlData, 0600)
}
