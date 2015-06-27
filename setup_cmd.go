package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func setupCmd(c *cli.Context) {
	log.Info("Starting setup")

	log.Info("Setup finished!")
}
