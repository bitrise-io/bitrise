package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
)

func main() {
	// Inputs
	var (
		versionFilePathParam = flag.String("file", "", `Version file path`)
	)

	flag.Parse()

	if versionFilePathParam == nil || *versionFilePathParam == "" {
		log.Fatalf(" [!] No version file parameter specified")
	}
	versionFilePath := *versionFilePathParam

	// Main
	versionFileBytes, err := ioutil.ReadFile(versionFilePath)
	if err != nil {
		log.Fatalf("Failed to read version file: %s", err)
	}
	versionFileContent := string(versionFileBytes)

	re := regexp.MustCompile(`const VERSION = "(?P<version>[0-9]+\.[0-9-]+\.[0-9-]+)"`)
	results := re.FindAllStringSubmatch(versionFileContent, -1)
	versionStr := ""
	for _, v := range results {
		versionStr = v[1]
	}
	if versionStr == "" {
		log.Fatalf("Failed to determine version")
	}

	fmt.Println(versionStr)
}
