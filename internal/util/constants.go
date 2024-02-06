package util

import (
	"fmt"
	"os"
)

var Eol = fmt.Sprintln()

const (
	DS              = string(os.PathSeparator)
	ConfigDir       = "." + DS + "config"
	ServerConfigFle = ConfigDir + DS + "servers.yml"
	BackupDir       = "." + DS + "backups"
	ConfigGenArg    = "gen"
)
