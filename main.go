package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/bitrise-io/bitrise/cli"
)

func main() {
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		print(err.Error())
	}

	cli.Run()
}
