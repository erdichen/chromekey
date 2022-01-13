package log

import (
	"log"

	"github.com/coreos/go-systemd/journal"
)

func Errorf(format string, v ...interface{}) {
	journal.Print(journal.PriAlert, format, v...)
	log.Printf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	journal.Print(journal.PriCrit, format, v...)
	log.Fatalf(format, v...)
}

func Infof(format string, v ...interface{}) {
	journal.Print(journal.PriInfo, format, v...)
	log.Printf(format, v...)
}
