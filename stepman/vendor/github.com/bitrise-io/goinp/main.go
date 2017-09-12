package main

import (
	"fmt"
	"log"

	"github.com/bitrise-io/goinp/goinp"
)

func main() {
	retStr, err := goinp.AskForString("Please enter some text here")
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Println("Entered text was:", retStr)

	retInt, err := goinp.AskForInt("Please enter a number")
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Println("Entered:", retInt)

	retBool, err := goinp.AskForBool("Yes or no?")
	if err != nil {
		log.Fatalln("Error:", err)
	}
	fmt.Println("Entered:", retBool)
}
