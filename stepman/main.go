package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/cli"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})
}

func main() {
	cli.Run()
}
