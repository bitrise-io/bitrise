package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/bitrise-io/bitrise/cli"
)

func main() {
	value, ok := os.LookupEnv("ENABLE_PROFILING")
	if ok && value == "1" {
		log.Println("pprof enabled over HTTP")
		go func() {
			log.Println(http.ListenAndServe("localhost:8080", nil))
		}()
	}

	cli.Run()
}
