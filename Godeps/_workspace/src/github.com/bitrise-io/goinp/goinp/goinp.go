package goinp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func AskForStringFromReader(messageToPrint string, inputReader io.Reader) (string, error) {
	scanner := bufio.NewScanner(inputReader)
	fmt.Printf("%s : ", messageToPrint)
	if scanner.Scan() {
		scannedText := scanner.Text()
		return scannedText, nil
	}
	return "", errors.New("Failed to get input - scanner failed.")
}

func AskForString(messageToPrint string) (string, error) {
	return AskForStringFromReader(messageToPrint, os.Stdin)
}

func AskForIntFromReader(messageToPrint string, inputReader io.Reader) (int64, error) {
	userInputStr, err := AskForStringFromReader(messageToPrint, inputReader)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(userInputStr, 10, 64)
}

func AskForInt(messageToPrint string) (int64, error) {
	return AskForIntFromReader(messageToPrint, os.Stdin)
}

func AskForBoolFromReader(messageToPrint string, inputReader io.Reader) (bool, error) {
	userInputStr, err := AskForStringFromReader(messageToPrint, inputReader)
	if err != nil {
		return false, err
	}
	lowercased := strings.ToLower(userInputStr)
	if lowercased == "yes" || lowercased == "y" {
		return true, nil
	}
	if lowercased == "no" || lowercased == "n" {
		return false, nil
	}
	return strconv.ParseBool(lowercased)
}

func AskForBool(messageToPrint string) (bool, error) {
	return AskForBoolFromReader(messageToPrint, os.Stdin)
}
