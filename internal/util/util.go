package util

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
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

func ParseEnvFromContent(b []byte, envEolChar byte) (map[string]string, bool) {
	if b == nil || len(b) == 0 {
		return nil, false
	}

	delim := "="
	var envMap = make(map[string]string)

	buf := bytes.NewBuffer(b)

	for {
		line, err := buf.ReadString(envEolChar)
		if err != nil {
			break
		}

		// remove all eol chars
		line = strings.Trim(line, string(envEolChar))
		line = strings.Trim(line, string('\r'))
		line = strings.Trim(line, string('\n'))

		// add k/v in map if found
		key, value, found := strings.Cut(line, delim)
		if !found || utf8.RuneCountInString(value) == 0 {
			continue
		}
		envMap[key] = value
	}

	// if nothing read
	if len(envMap) == 0 {
		return nil, false
	}

	return envMap, true
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

func ReadStream(stream *io.Reader, readTo *[]byte) {
	ReadLinesFromStream(stream, func(b []byte) {
		*readTo = append(*readTo, b...)
	})
}

func ReadLinesFromStream(stream *io.Reader, cb func(b []byte)) {
	scanner := bufio.NewScanner(*stream)
	for scanner.Scan() {
		cb(scanner.Bytes())
	}
}

func CurrentTimeStr() string {
	return time.Now().Format(time.DateTime)
}

// CreatePath creates path(dirs) recursively if not exist
// when provided path is a file path (@pathIsFile = true) then creates all dirs except the file
func CreatePath(path string, perm os.FileMode, pathIsFile bool) error {
	parentDir := path
	if pathIsFile {
		parentDir = filepath.Dir(path)
	}

	// do not try to create dir if that is not a nested path
	emptyDir := []string{".", "..", "/", string(os.PathSeparator)}
	for _, d := range emptyDir {
		if d == parentDir {
			return nil
		}
	}

	return os.MkdirAll(parentDir, perm)
}
