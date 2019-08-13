package envutil

import (
	"fmt"
	"os"
)

// RevokableSetenv ...
func RevokableSetenv(envKey, envValue string) (func() error, error) {
	origValue := os.Getenv(envKey)
	revokeFn := func() error {
		return os.Setenv(envKey, origValue)
	}

	return revokeFn, os.Setenv(envKey, envValue)
}

// SetenvForFunction ...
func SetenvForFunction(envKey, envValue string, fn func()) error {
	revokeFn, err := RevokableSetenv(envKey, envValue)
	if err != nil {
		return err
	}

	fn()

	return revokeFn()
}

// StringFlagOrEnv - returns the value of the flag if specified, otherwise the env's value.
// Empty string counts as not specified!
func StringFlagOrEnv(flagValue *string, envKey string) string {
	if flagValue != nil && *flagValue != "" {
		return *flagValue
	}
	return os.Getenv(envKey)
}

// GetenvWithDefault - returns the env if specified, default value otherwise
func GetenvWithDefault(envKey, defValue string) string {
	retVal := os.Getenv(envKey)
	if retVal == "" {
		return defValue
	}
	return retVal
}

// RequiredEnv - returns the env's value if specified, otherwise it returns an error that explains the key is required.
// Use this function to reduce error prone code duplication.
// E.g. instead of doing this in your code:
//
// 	myVar1 := os.Getenv("MY_ENV1")
// 	if len(myVar1) < 1 {
// 		return nil, errors.New("MY_ENV1 required")
// 	}
//
// You can use this function like:
//
// 	myVar1, err := requiredEnv("MY_ENV1")
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// In the first example you have to specify myVar1 and MY_ENV1 two times, which can lead to
// issues if you copy-paste that code but e.g. forget to change the var name in the
// 	if len(myVar1) < 1
// line, or if you forget to change the var key in the error message/string.
func RequiredEnv(envKey string) (string, error) {
	if val := os.Getenv(envKey); len(val) > 0 {
		return val, nil
	}
	return "", fmt.Errorf("required environment variable (%s) not provided", envKey)
}
