package logs

import (
	"log"
	"os"
)

var logger *log.Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

func SetPrefix(prefix string) {
	logger = log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile)
}

func Logger() *log.Logger {
	return logger
}

func Printf(format string, a ...any) {
	logger.Printf(format, a...)
}
