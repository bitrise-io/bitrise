package utils

import (
	"os"
	"os/exec"
	"strings"
)

// CheckProgramInstalledPath ...
func CheckProgramInstalledPath(clcommand string) (string, error) {
	cmd := exec.Command("which", clcommand)
	cmd.Stderr = os.Stderr
	outBytes, err := cmd.Output()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// YAMLToJSONKeyTypeConversion ...
func YAMLToJSONKeyTypeConversion(i interface{}) interface{} {
	switch x := i.(type) {
	case map[string]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k] = YAMLToJSONKeyTypeConversion(v)
		}
		return m2
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = YAMLToJSONKeyTypeConversion(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = YAMLToJSONKeyTypeConversion(v)
		}
	}
	return i
}
