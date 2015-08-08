package bitrise

import (
	"os"
	"path"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
)

func TestSetupForVersionChecks(t *testing.T) {
	currPth, err := pathutil.CurrentWorkingDirectoryAbsolutePath()
	if err != nil {
		t.Fatal("Failed to get curr abs path: ", err)
	}

	fakeHomePth := path.Join(currPth, "_FAKE_HOME")
	err = os.Mkdir(fakeHomePth, 0777)
	if err != nil {
		t.Fatal("Failed to create fake HOME: ", err)
	}
	defer func() {
		err := os.RemoveAll(fakeHomePth)
		if err != nil {
			t.Error("Failed to remove FAKE HOME: ", err)
		}
	}()
	err = os.Setenv("HOME", fakeHomePth)
	if err != nil {
		t.Fatal("Failed to set (fake) HOME: ", err)
	}

	t.Log("First check - should be empty and so should fail")
	if isOK := CheckIsSetupWasDoneForVersion("0.9.7"); isOK {
		t.Fatal("Should not be OK")
	}

	t.Log("Write an 'ok' for the first test version")
	err = SaveSetupSuccessForVersion("0.9.7")
	if err != nil {
		t.Fatal("Error: ", err)
	}

	t.Log("Check for the 'ok' version - should be ok")
	if isOK := CheckIsSetupWasDoneForVersion("0.9.7"); !isOK {
		t.Fatal("Should be OK")
	}

	t.Log("Check for a newer version - should NOT be ok")
	if isOK := CheckIsSetupWasDoneForVersion("0.9.8"); isOK {
		t.Fatal("Should NOT be OK")
	}
}
