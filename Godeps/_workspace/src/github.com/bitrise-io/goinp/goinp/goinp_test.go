package goinp

import (
	"strings"
	"testing"
)

func TestAskForStringFromReader(t *testing.T) {
	t.Log("TestAskForString")

	testUserInput := "this is some text"

	res, err := AskForStringFromReader("Enter some text", strings.NewReader(testUserInput))
	if err != nil {
		t.Fatal(err)
	}
	if res != testUserInput {
		t.Fatalf("Scanned input (%s) does not match expected (%s)", res, testUserInput)
	}
}

func TestAskForIntFromReader(t *testing.T) {
	t.Log("TestAskForString")

	testUserInput := "31"

	res, err := AskForIntFromReader("Enter a number", strings.NewReader(testUserInput))
	if err != nil {
		t.Fatal(err)
	}
	if res != 31 {
		t.Fatalf("Scanned input (%s) does not match expected (%s)", res, testUserInput)
	}
}

func TestAskForBoolFromReader(t *testing.T) {
	t.Log("TestAskForString")

	// yes
	testUserInput := "y"
	res, err := AskForBoolFromReader("Yes or no?", strings.NewReader(testUserInput))
	if err != nil {
		t.Fatal(err)
	}
	if res != true {
		t.Fatalf("Scanned input (%s) does not match expected (%s)", res, testUserInput)
	}

	// no
	testUserInput = "no"
	res, err = AskForBoolFromReader("Yes or no?", strings.NewReader(testUserInput))
	if err != nil {
		t.Fatal(err)
	}
	if res != false {
		t.Fatalf("Scanned input (%s) does not match expected (%s)", res, testUserInput)
	}
}

func TestParseBool(t *testing.T) {
	t.Log("Simple Yes")
	testUserInput := "y"
	isYes, err := ParseBool("YeS")
	if err != nil {
		t.Fatal(err)
	}
	if !isYes {
		t.Fatalf("Scanned input (%s) does not match expected (%s)", isYes, testUserInput)
	}

	t.Log("Simple No")
	testUserInput = "no"
	isYes, err = ParseBool("n")
	if err != nil {
		t.Fatal(err)
	}
	if isYes {
		t.Fatalf("Scanned input (%s) does not match expected (%s)", isYes, testUserInput)
	}

	t.Log("Newline in yes - trim")
	testUserInput = `
 yes
`
	isYes, err = ParseBool(testUserInput)
	if err != nil {
		t.Fatal(err)
	}
	if !isYes {
		t.Fatalf("Scanned input (%s) does not match expected (%s)", isYes, testUserInput)
	}
}
