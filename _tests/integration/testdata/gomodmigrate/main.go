package main

import (
	"fmt"
	"os"

	shellquote "github.com/kballard/go-shellquote"
)

func main() {
	str := "Hello World!"
	arr, err := shellquote.Split(str)
	if err != nil {
		fmt.Printf("failed to split: %v", err)
		os.Exit(1)
	}
	for _, str := range arr {
		fmt.Println(str)
	}
	os.Exit(0)
}
