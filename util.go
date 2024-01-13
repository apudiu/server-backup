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

func genProjectInfo(projectName string) projectInfo {
	info := projectInfo{}

	// clean name
	replaceChars := map[string]string{
		"\n": "",
		"$":  "",
		"#":  "",
		",":  "",
		";":  "",
		"[":  "",
		"]":  "",
		"{":  "",
		"}":  "",
		"(":  "",
		")":  "",
		"~":  "",
		" ":  "_",
		".":  "-",
	}

	cn := projectName
	for replace, with := range replaceChars {
		cn = strings.ReplaceAll(cn, replace, with)
	}

	p := fmt.Sprintf("./%s/", cn)
	caCertNamePathWithoutExt := p + cn + "_ca_cert"
	srvCertNamePathWithoutExt := p + cn + "_srv_cert"

	info.Name = cn                                       // ex: domain-com
	info.Path = p                                        // ex: ./domain-com/
	info.CaKey = p + cn + "_ca_key.pem"                  // ex: ./domain-com/_ca_key.pem
	info.CaCertName = caCertNamePathWithoutExt           // ex: ./domain-com/_ca_cert
	info.CaCertDer = caCertNamePathWithoutExt + ".crt"   // ex: ./domain-com/_ca_cert.crt
	info.CaCertPem = caCertNamePathWithoutExt + ".pem"   // ex: ./domain-com/_ca_cert.pem
	info.SrvCertName = srvCertNamePathWithoutExt         // ex: ./domain-com/_srv_cert
	info.SrvKey = p + cn + "_srv_key.pem"                // ex: ./domain-com/_srv_key.pem
	info.SrvCertDer = srvCertNamePathWithoutExt + ".crt" // ex: ./domain-com/_srv_cert.crt
	info.SrvCertPem = srvCertNamePathWithoutExt + ".pem" // ex: ./domain-com/_srv_cert.pem
	info.SrvCertPfx = srvCertNamePathWithoutExt + ".pfx" // ex: ./domain-com/_srv_cert.pfx

	return info
}
