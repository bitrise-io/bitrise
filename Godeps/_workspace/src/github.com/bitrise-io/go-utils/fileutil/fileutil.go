package fileutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bitrise-io/go-utils/pathutil"
)

// WriteStringToFile ...
func WriteStringToFile(pth string, fileCont string) error {
	return WriteBytesToFile(pth, []byte(fileCont))
}

// WriteBytesToFile ...
func WriteBytesToFile(pth string, fileCont []byte) error {
	if pth == "" {
		return errors.New("No path provided")
	}

	file, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println(" [!] Failed to close file:", err)
		}
	}()

	if _, err := file.Write(fileCont); err != nil {
		return err
	}

	return nil
}

// AppendStringToFile ...
func AppendStringToFile(pth string, fileCont string) error {
	return AppendBytesToFile(pth, []byte(fileCont))
}

// AppendBytesToFile ...
func AppendBytesToFile(pth string, fileCont []byte) error {
	if pth == "" {
		return errors.New("No path provided")
	}

	var file *os.File
	filePerm, err := GetFilePermissions(pth)
	if err != nil {
		// create the file
		file, err = os.Create(pth)
	} else {
		// open for append
		file, err = os.OpenFile(pth, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerm)
	}
	if err != nil {
		// failed to create or open-for-append the file
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println(" [!] Failed to close file:", err)
		}
	}()

	if _, err := file.Write(fileCont); err != nil {
		return err
	}

	return nil
}

// ReadBytesFromFile ...
func ReadBytesFromFile(pth string) ([]byte, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return []byte{}, err
	} else if !isExists {
		return []byte{}, errors.New(fmt.Sprint("No file found at path", pth))
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

// ReadStringFromFile ...
func ReadStringFromFile(pth string) (string, error) {
	contBytes, err := ReadBytesFromFile(pth)
	if err != nil {
		return "", err
	}
	return string(contBytes), nil
}

// GetFilePermissions ...
func GetFilePermissions(filePth string) (os.FileMode, error) {
	info, err := os.Stat(filePth)
	if err != nil {
		return 0, err
	}
	mode := info.Mode()
	return mode, nil
}
