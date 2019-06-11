package goinp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/bitrise-io/go-utils/log"
)

//=======================================
// String
//=======================================

// AskForStringFromReaderWithDefault ...
func AskForStringFromReaderWithDefault(messageToPrint, defaultValue string, inputReader io.Reader) (string, error) {
	scanner := bufio.NewScanner(inputReader)

	if defaultValue == "" {
		fmt.Printf("%s : ", messageToPrint)
	} else {
		fmt.Printf("%s [%s] : ", messageToPrint, defaultValue)
	}

	scannedText := ""
	if scanner.Scan() {
		scannedText = scanner.Text()
		scannedText = strings.TrimRight(scannedText, " ")
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to get input - scanner failed wit error: %s", err)
	}

	if scannedText == "" {
		if defaultValue != "" {
			return defaultValue, nil
		}
		return "", errors.New("failed to get input - scanner failed")
	}

	return scannedText, nil
}

// AskForStringFromReader ...
func AskForStringFromReader(messageToPrint string, inputReader io.Reader) (string, error) {
	return AskForStringFromReaderWithDefault(messageToPrint, "", inputReader)
}

// AskForStringWithDefault ...
func AskForStringWithDefault(messageToPrint, defaultValue string) (string, error) {
	return AskForStringFromReaderWithDefault(messageToPrint, defaultValue, os.Stdin)
}

// AskForString ...
func AskForString(messageToPrint string) (string, error) {
	return AskForStringFromReader(messageToPrint, os.Stdin)
}

func writeStdinClearable(text string) error {
	for _, c := range []byte(text) {
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCSTI, uintptr(unsafe.Pointer(&c))); errno != 0 {
			return fmt.Errorf("failed to write to stdin, err no: %d", errno)
		}
	}
	return nil
}

// AskOptions ...
func AskOptions(title string, defaultValue string, optional bool, options ...string) (string, error) {
	const customValueOptionText = "<custom value>"

	if len(options) == 0 {
		for {
			if title != "" {
				fmt.Print("Enter value for \"" + strings.TrimSuffix(title, ":") + "\": ")
			}
			var (
				input string
				err   error
			)

			if defaultValue != "" {
				if err := writeStdinClearable(defaultValue); err != nil {
					return "", err
				}
			}

			input, err = bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return "", err
			}

			if !optional && strings.TrimSpace(input) == "" {
				log.Errorf("value must be specified")
				continue
			}

			return strings.TrimSpace(input), nil
		}
	}

	// add last option if optional so user can decide to input value manually
	if optional && len(options) > 0 {
		options = append(options, customValueOptionText)
	}

	// selector with one option -> auto select
	if len(options) == 1 {
		return options[0], nil
	}

	if title != "" {
		fmt.Println("Select \"" + strings.TrimSuffix(title, ":") + "\" from the list:")
	}

	for i, option := range options {
		log.Printf("[%d] : %s", i+1, option)
	}

	for {
		fmt.Print("Type in the option's number, then hit Enter: ")

		answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Errorf("failed to read input value")
			continue
		}

		optionNo, err := strconv.Atoi(strings.TrimSpace(answer))
		if err != nil {
			log.Errorf("failed to parse option number, pick a number from 1-%d", len(options))
			continue
		}

		if optionNo-1 < 0 || optionNo-1 >= len(options) {
			log.Errorf("invalid option number, pick a number 1-%d", len(options))
			continue
		}

		if options[optionNo-1] == customValueOptionText {
			return AskOptions(title, "", true)
		}

		return options[optionNo-1], nil
	}
}

//=======================================
// Path
//=======================================

// AskForPathFromReaderWithDefault asks for a path. The difference between this
//  and the generic "AskForString..." functions is that this'll
//  clean up the input. For example, if the user drag-and-drops a file/dir
//  for the input then the input might include back-slash escapes for
//  spaces in the path - these will be removed, so the
//  returned path will be "path/with space" instead of "path/with\ space".
func AskForPathFromReaderWithDefault(messageToPrint, defaultValue string, inputReader io.Reader) (string, error) {
	str, err := AskForStringFromReaderWithDefault(messageToPrint, defaultValue, inputReader)
	if err != nil {
		return "", err
	}

	return strings.Replace(str, "\\", "", -1), nil
}

// AskForPathFromReader ...
func AskForPathFromReader(messageToPrint string, inputReader io.Reader) (string, error) {
	return AskForPathFromReaderWithDefault(messageToPrint, "", inputReader)
}

// AskForPathWithDefault ...
func AskForPathWithDefault(messageToPrint, defaultValue string) (string, error) {
	return AskForPathFromReaderWithDefault(messageToPrint, defaultValue, os.Stdin)
}

// AskForPath ...
func AskForPath(messageToPrint string) (string, error) {
	return AskForPathFromReader(messageToPrint, os.Stdin)
}

//=======================================
// Int
//=======================================

// AskForIntFromReaderWithDefault ...
func AskForIntFromReaderWithDefault(messageToPrint string, defaultValue int, inputReader io.Reader) (int64, error) {
	userInputStr, err := AskForStringFromReaderWithDefault(messageToPrint, fmt.Sprintf("%d", defaultValue), inputReader)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(userInputStr, 10, 64)
}

// AskForIntFromReader ...
func AskForIntFromReader(messageToPrint string, inputReader io.Reader) (int64, error) {
	userInputStr, err := AskForStringFromReader(messageToPrint, inputReader)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(userInputStr, 10, 64)
}

// AskForIntWithDeafult ...
func AskForIntWithDeafult(messageToPrint string, defaultValue int) (int64, error) {
	return AskForIntFromReaderWithDefault(messageToPrint, defaultValue, os.Stdin)
}

// AskForInt ...
func AskForInt(messageToPrint string) (int64, error) {
	return AskForIntFromReader(messageToPrint, os.Stdin)
}

//=======================================
// Bool
//=======================================

// ParseBool ...
func ParseBool(userInputStr string) (bool, error) {
	if userInputStr == "" {
		return false, errors.New("no string to parse")
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

// AskForBoolFromReaderWithDefaultValue ...
func AskForBoolFromReaderWithDefaultValue(messageToPrint string, defaultValue bool, inputReader io.Reader) (bool, error) {
	keywordYes := "yes"
	keywordNo := "no"
	if defaultValue == true {
		keywordYes = "YES"
	} else {
		keywordNo = "NO"
	}
	fmt.Printf("%s [%s/%s]: ", messageToPrint, keywordYes, keywordNo)

	scanner := bufio.NewScanner(inputReader)
	scannedText := ""
	if scanner.Scan() {
		scannedText = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to get input - scanner failed wit error: %s", err)
	}

	if scannedText == "" {
		return defaultValue, nil
	}
	return ParseBool(scannedText)
}

// AskForBoolFromReader ...
func AskForBoolFromReader(messageToPrint string, inputReader io.Reader) (bool, error) {
	userInputStr, err := AskForStringFromReader(messageToPrint+" [yes/no]", inputReader)
	if err != nil {
		return false, err
	}

	return ParseBool(userInputStr)
}

// AskForBoolWithDefault ...
func AskForBoolWithDefault(messageToPrint string, defaultValue bool) (bool, error) {
	return AskForBoolFromReaderWithDefaultValue(messageToPrint, defaultValue, os.Stdin)
}

// AskForBool ...
func AskForBool(messageToPrint string) (bool, error) {
	return AskForBoolFromReader(messageToPrint, os.Stdin)
}

//=======================================
// Select
//=======================================

// SelectFromStringsFromReaderWithDefault ...
func SelectFromStringsFromReaderWithDefault(messageToPrint string, defaultValue int, options []string, inputReader io.Reader) (string, error) {
	fmt.Printf("%s\n", messageToPrint)
	fmt.Println("Please select from the list:")
	for idx, anOption := range options {
		fmt.Printf("[%d] : %s\n", idx+1, anOption)
	}

	selectedOptionNum, err := AskForIntFromReaderWithDefault("(type in the option's number, then hit Enter)", defaultValue, inputReader)
	if err != nil {
		return "", err
	}

	if selectedOptionNum < 1 {
		return "", fmt.Errorf("invalid option: You entered a number less than 1")
	}
	if selectedOptionNum > int64(len(options)) {
		return "", fmt.Errorf("invalid option: You entered a number greater than the last option's number")
	}
	return options[selectedOptionNum-1], nil
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
		return "", fmt.Errorf("invalid option: You entered a number less than 1")
	}
	if selectedOptionNum > int64(len(options)) {
		return "", fmt.Errorf("invalid option: You entered a number greater than the last option's number")
	}
	return options[selectedOptionNum-1], nil
}

// SelectFromStringsWithDefault ...
func SelectFromStringsWithDefault(messageToPrint string, defaultValue int, options []string) (string, error) {
	return SelectFromStringsFromReaderWithDefault(messageToPrint, defaultValue, options, os.Stdin)
}

// SelectFromStrings ...
func SelectFromStrings(messageToPrint string, options []string) (string, error) {
	return SelectFromStringsFromReader(messageToPrint, options, os.Stdin)
}
