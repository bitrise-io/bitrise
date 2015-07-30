package envman

import (
	"os"
	"strings"
	"testing"

	"github.com/bitrise-io/envman/models"
)

func TestExpandEnvsInString(t *testing.T) {
	t.Log("TestExpandEnvsInString")

	inp := "${HOME} home"
	expanded := expandEnvsInString(inp)

	home := os.Getenv("HOME")
	if home != "" {
		should := home + " home"
		if expanded != should {
			t.Fatalf("Incorrect expand (%s), should be: (%s)", expanded, should)
		}
	} else {
		t.Fatal("No ${HOME} env set")
	}
}

func TestCommandEnvs(t *testing.T) {
	t.Log("TestCommandEnvs")

	env1 := models.EnvironmentItemModel{
		"test_key1": "test_value1",
	}
	err := env1.FillMissingDefaults()
	if err != nil {
		t.Fatal(err)
	}

	env2 := models.EnvironmentItemModel{
		"test_key2": "test_value2",
	}
	err = env2.FillMissingDefaults()
	if err != nil {
		t.Fatal(err)
	}
	envs := []models.EnvironmentItemModel{env1, env2}

	sessionEnvs, err := commandEnvs(envs)
	if err != nil {
		t.Fatal(err)
	}

	env1Found := false
	env2Found := false
	for _, envString := range sessionEnvs {
		comp := strings.Split(envString, "=")
		key := comp[0]
		value := comp[1]

		envKey1, envValue1, err := env1.GetKeyValuePair()
		if err != nil {
			t.Fatal(err)
		}

		envKey2, envValue2, err := env2.GetKeyValuePair()
		if err != nil {
			t.Fatal(err)
		}

		if key == envKey1 && value == envValue1 {
			env1Found = true
		}
		if key == envKey2 && value == envValue2 {
			env2Found = true
		}
	}
	if env1Found == false {
		t.Fatalf("Failed to set env (%v)", env1)
	}
	if env2Found == false {
		t.Fatalf("Failed to set env (%v)", env2)
	}
}

// func TestRunCmd(t *testing.T) {
// 	t.Log("TestRunCmd")
// }
