package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/cli"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})
	log.SetOutput(os.Stdout)
}

func main() {
	cli.Run()
}
