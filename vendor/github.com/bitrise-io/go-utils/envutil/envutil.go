package envutil

import "os"

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
