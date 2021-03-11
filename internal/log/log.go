package log

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type LoggerInterface interface {
	Print(...interface{})
	Println(...interface{})
	Printf(string, ...interface{})
	Fatal(...interface{})
	Fatalln(...interface{})
}

var (
	Debug LoggerInterface
	Info  LoggerInterface
	Error LoggerInterface
)

// loggers init
func init() {
	debugHandle := ioutil.Discard
	if value := strings.TrimSpace(os.Getenv("DEBUG")); value == "1" {
		debugHandle = os.Stdout
	}
	Debug = log.New(debugHandle, "DEBUG: ", log.LstdFlags)
	Info = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	Error = log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	Debug.Println("Logging initialized")
}
