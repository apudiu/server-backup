package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func writeToFile(path string, content []byte, perm os.FileMode) {
	err := os.WriteFile(path, content, perm)
	failIfErr(err, "Failed to write "+path)
}

func readFromFile(path string) []byte {
	fileBytes, err := os.ReadFile(path)
	failIfErr(err, "File "+path+" read error: ")
	return fileBytes
}

func failIfErr(e error, prependMsg ...string) {
	if e != nil {
		log.Fatalf("%s, Error: %s \n", prependMsg[0], e.Error())
	}
}

func isPathExist(p string) (bool, error) {
	if _, err := os.Stat(p); err != nil {
		return false, err
	}
	return true, nil
}

func readUserInput(msg string) (val string, err error) {
	fmt.Print(msg + ": ")
	r := bufio.NewReader(os.Stdin)

	val, err = r.ReadString('\n')
	val = strings.Trim(val, "\n")
	val = strings.Trim(val, "\r")
	val = strings.Trim(val, " ")
	return
}

func parseEnv(path string) map[string]string {
	f, err := os.Open(path)
	failIfErr(err, "Failed to read env file from "+path)
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
