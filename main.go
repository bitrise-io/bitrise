package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/bitrise-io/bitrise/cli"
)

func main() {
	_ = http.ListenAndServe("localhost:8080", nil)

	cli.Run()
}
