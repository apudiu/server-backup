package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

var configFilePath = "./config.yml"

type caInfo struct {
	Organization  string `yaml:"organization"`
	Country       string `yaml:"country"`
	Locality      string `yaml:"city"`
	StreetAddress string `yaml:"street"`
	PostalCode    string `yaml:"postalCode"`
}

type caConfig struct {
	Key  string `yaml:"privateKeyPath"`
	Cert string `yaml:"certificatePath"`

	Serial      int `yaml:"serial"`
	ExpiryYears int `yaml:"expiryYears"`
	Info        caInfo
}

type serverInfo struct {
	Organization  string   `yaml:"organization"`
	Country       string   `yaml:"country"`
	Locality      string   `yaml:"city"`
	StreetAddress string   `yaml:"street"`
	PostalCode    string   `yaml:"postalCode"`
	CommonNames   []string `yaml:"domainNames"`
}

type serverConfig struct {
	Key string `yaml:"privateKeyPath"`

	Serial      int `yaml:"serial"`
	ExpiryYears int `yaml:"expiryYears"`
	Info        serverInfo
	IpAddresses []string `yaml:"ipAddresses"`
	// encryption password
	Password string `yaml:"password"`
}

type projectInfo struct {
	Name string
	Path string
	// relative path to the files
	CaKey       string
	CaCertName  string
	CaCertDer   string
	CaCertPem   string
	SrvKey      string
	SrvCertName string
	SrvCertDer  string
	SrvCertPem  string
	SrvCertPfx  string
}

type config struct {
	// will be used as dir & file name
	ProjectName string      `yaml:"projectName"`
	ProjectInfo projectInfo `yaml:"-"`
	Ca          caConfig
	Server      serverConfig
}

func (c *config) Parse() {
	if exists, _ := isPathExist(configFilePath); !exists {
		log.Fatalln("Config file unavailable at " + configFilePath)
	}

	fb, err := os.ReadFile(configFilePath)
	failIfErr(err)

	unmarshalErr := yaml.Unmarshal(fb, c)
	failIfErr(unmarshalErr)

	// attach generated info
	c.ProjectInfo = genProjectInfo(c.ProjectName)
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
	pn := "example.com"
	pi := genProjectInfo(pn)

	c := config{
		ProjectName: pn,
		Ca: caConfig{
			Key:  pi.CaKey,
			Cert: pi.CaCertPem,

			Serial:      2020,
			ExpiryYears: 10,
			Info: caInfo{
				Organization:  "Snebtaf",
				Country:       "UK",
				Locality:      "London",
				StreetAddress: "Bricklane",
				PostalCode:    "E1 6QL",
			},
		},
		Server: serverConfig{
			Key: pi.SrvKey,

			Serial:      2021,
			ExpiryYears: 5,
			Info: serverInfo{
				Organization:  "Ordering2online",
				Country:       "BD",
				Locality:      "Sylhet",
				StreetAddress: "Ambarkhana",
				PostalCode:    "1201",
				CommonNames:   []string{"print.digitafact.com"},
			},
			IpAddresses: []string{"192.168.0.121"},
			Password:    "1234",
		},
	}

	ymlData, err := yaml.Marshal(&c)
	failIfErr(err)

	writeToFile(configFilePath, ymlData, 0600)
}
