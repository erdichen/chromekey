package log

import (
	"log"
	"os"

	"github.com/coreos/go-systemd/journal"
)

var isTerm = false

func init() {
	fileInfo, _ := os.Stdout.Stat()
	isTerm = (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func Errorf(format string, v ...interface{}) {
	if !isTerm {
		journal.Print(journal.PriAlert, format, v...)
	}
	log.Printf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	if !isTerm {
		journal.Print(journal.PriCrit, format, v...)
	} else {
		log.Fatalf(format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if !isTerm {
		journal.Print(journal.PriInfo, format, v...)
	} else {
		log.Printf(format, v...)
	}
}
