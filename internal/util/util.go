package util

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func WriteToFile(path string, content []byte, perm os.FileMode) {
	err := os.WriteFile(path, content, perm)
	FailIfErr(err, "Failed to write "+path)
}

func ReadFromFile(path string) []byte {
	fileBytes, err := os.ReadFile(path)
	FailIfErr(err, "File "+path+" read error: ")
	return fileBytes
}

func FailIfErr(e error, prependMsg ...string) {
	if e != nil {
		log.Fatalf("%s, Error: %s \n", prependMsg[0], e.Error())
	}
}

func IsPathExist(p string) (bool, error) {
	if _, err := os.Stat(p); err != nil {
		return false, err
	}
	return true, nil
}

func ReadUserInput(msg string) (val string, err error) {
	fmt.Print(msg + ": ")
	r := bufio.NewReader(os.Stdin)

	val, err = r.ReadString('\n')
	val = strings.Trim(val, "\n")
	val = strings.Trim(val, "\r")
	val = strings.Trim(val, " ")
	return
}

func ParseEnv(path string) map[string]string {
	f, err := os.Open(path)
	FailIfErr(err, "Failed to read env file from "+path)
	defer f.Close()

	delim := "="
	var envMap = make(map[string]string)

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		l := scanner.Text()
		if valid := strings.Contains(l, delim); !valid {
			continue
		}

		s := strings.Split(l, delim)
		envMap[s[0]] = s[1]
	}

	fmt.Println(envMap)

	return envMap
}

// KeyGetFromFile parses pe encoded private key form disk
func KeyGetFromFile(path string) (key *rsa.PrivateKey, pemBytes []byte, err error) {
	pemBytes = ReadFromFile(path)
	keyBlock, _ := pem.Decode(pemBytes)
	if keyBlock == nil || keyBlock.Type != "RSA PRIVATE KEY" {
		err = errors.New("failed to decode PEM block containing private key from " + path)
		return
	}

	key, keyParseErr := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if keyParseErr != nil {
		err = errors.New("failed to parse private key from " + path)
		return
	}

	return key, pemBytes, nil
}

/*
MakeAbsoluteFilePath constructs platform independent absolute file path
MakeAbsoluteFilePath("home", "user") // /home/user
*/
func MakeAbsoluteFilePath(paths ...string) (fullPath string) {
	fullPath = string(os.PathSeparator)
	fullPath += filepath.Join(paths...)
	return
}

func ErrWithPrefix(msg string, e error) error {
	return errors.New(msg + " - " + e.Error())
}

func EolChar() string {
	var PS = fmt.Sprintf("%v", os.PathSeparator)
	var lb = "\n"

	if PS != "/" {
		lb = "\r\n"
	}
	return lb
}
