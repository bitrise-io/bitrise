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

// AskForStringFromReader ...
func AskForStringFromReader(messageToPrint string, inputReader io.Reader) (string, error) {
	scanner := bufio.NewScanner(inputReader)
	fmt.Printf("%s : ", messageToPrint)
	if scanner.Scan() {
		scannedText := scanner.Text()
		return strings.TrimRight(scannedText, " "), nil
	}
	return "", errors.New("Failed to get input - scanner failed.")
}

// AskForString ...
func AskForString(messageToPrint string) (string, error) {
	return AskForStringFromReader(messageToPrint, os.Stdin)
}

// AskForIntFromReader ...
func AskForIntFromReader(messageToPrint string, inputReader io.Reader) (int64, error) {
	userInputStr, err := AskForStringFromReader(messageToPrint, inputReader)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(userInputStr, 10, 64)
}

// AskForInt ...
func AskForInt(messageToPrint string) (int64, error) {
	return AskForIntFromReader(messageToPrint, os.Stdin)
}

// ParseBool ...
func ParseBool(userInputStr string) (bool, error) {
	if userInputStr == "" {
		return false, errors.New("No string to parse")
	}
	userInputStr = strings.TrimSpace(userInputStr)

	lowercased := strings.ToLower(userInputStr)
	if lowercased == "yes" || lowercased == "y" {
		return true, nil
	}
	if lowercased == "no" || lowercased == "n" {
		return false, nil
	}
	return strconv.ParseBool(lowercased)
}

// AskForBoolFromReader ...
func AskForBoolFromReader(messageToPrint string, inputReader io.Reader) (bool, error) {
	userInputStr, err := AskForStringFromReader(messageToPrint+" [yes/no]", inputReader)
	if err != nil {
		return false, err
	}

	return ParseBool(userInputStr)
}

// AskForBool ...
func AskForBool(messageToPrint string) (bool, error) {
	return AskForBoolFromReader(messageToPrint, os.Stdin)
}

// SelectFromStringsFromReader ...
func SelectFromStringsFromReader(messageToPrint string, options []string, inputReader io.Reader) (string, error) {
	fmt.Printf("%s\n", messageToPrint)
	fmt.Println("Please select from the list:")
	for idx, anOption := range options {
		fmt.Printf("[%d] : %s\n", idx+1, anOption)
	}
	selectedOptionNum, err := AskForIntFromReader("(type in the option's number, then hit Enter)", inputReader)
	if err != nil {
		return "", err
	}
	if selectedOptionNum < 1 {
		return "", fmt.Errorf("Invalid option: You entered a number less than 1")
	}
	if selectedOptionNum > int64(len(options)) {
		return "", fmt.Errorf("Invalid option: You entered a number greater than the last option's number")
	}
	return options[selectedOptionNum-1], nil
}

// SelectFromStrings ...
func SelectFromStrings(messageToPrint string, options []string) (string, error) {
	return SelectFromStringsFromReader(messageToPrint, options, os.Stdin)
}
