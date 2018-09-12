package cli

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/urfave/cli"
)

var errTimeout = errors.New("timeout")

func envListSizeInBytes(envs []models.EnvironmentItemModel) (int, error) {
	valueSizeInBytes := 0
	for _, env := range envs {
		_, value, err := env.GetKeyValuePair()
		if err != nil {
			return 0, err
		}
		valueSizeInBytes += len([]byte(value))
	}
	return valueSizeInBytes, nil
}

func validateEnv(key, value string, envList []models.EnvironmentItemModel) (string, error) {
	if key == "" {
		return "", errors.New("Key is not specified, required")
	}

	configs, err := envman.GetConfigs()
	if err != nil {
		return "", err
	}

	valueSizeInBytes := len([]byte(value))
	if configs.EnvBytesLimitInKB > 0 {
		if valueSizeInBytes > configs.EnvBytesLimitInKB*1024 {
			valueSizeInKB := ((float64)(valueSizeInBytes)) / 1024.0
			log.Warnf("environment value (%s...) too large", value[0:100])
			log.Warnf("environment value size (%#v KB) - max allowed size: %#v KB", valueSizeInKB, (float64)(configs.EnvBytesLimitInKB))
			return "environment value too large - rejected", nil
		}
	}

	if configs.EnvListBytesLimitInKB > 0 {
		envListSizeInBytes, err := envListSizeInBytes(envList)
		if err != nil {
			return "", err
		}
		if envListSizeInBytes+valueSizeInBytes > configs.EnvListBytesLimitInKB*1024 {
			listSizeInKB := (float64)(envListSizeInBytes)/1024 + (float64)(valueSizeInBytes)/1024
			log.Warn("environment list too large")
			log.Warnf("environment list size (%#v KB) - max allowed size: %#v KB", listSizeInKB, (float64)(configs.EnvListBytesLimitInKB))
			return "", errors.New("environment list too large")
		}
	}
	return value, nil
}

func addEnv(key string, value string, expand, replace, skipIfEmpty bool) error {
	// Load envs, or create if not exist
	environments, err := envman.ReadEnvsOrCreateEmptyList()
	if err != nil {
		return err
	}

	// Validate input
	validatedValue, err := validateEnv(key, value, environments)
	if err != nil {
		return err
	}
	value = validatedValue

	// Add or update envlist
	newEnv := models.EnvironmentItemModel{
		key: value,
		models.OptionsKey: models.EnvironmentItemOptionsModel{
			IsExpand:    pointers.NewBoolPtr(expand),
			SkipIfEmpty: pointers.NewBoolPtr(skipIfEmpty),
		},
	}
	if err := newEnv.NormalizeValidateFillDefaults(); err != nil {
		return err
	}

	newEnvSlice, err := envman.UpdateOrAddToEnvlist(environments, newEnv, replace)
	if err != nil {
		return err
	}

	return envman.WriteEnvMapToFile(envman.CurrentEnvStoreFilePath, newEnvSlice)
}

func loadValueFromFile(pth string) (string, error) {
	buf, err := ioutil.ReadFile(pth)
	if err != nil {
		return "", err
	}

	str := string(buf)
	return str, nil
}

func logEnvs() error {
	environments, err := envman.ReadEnvs(envman.CurrentEnvStoreFilePath)
	if err != nil {
		return err
	}

	if len(environments) == 0 {
		log.Info("[ENVMAN] Empty envstore")
	} else {
		for _, env := range environments {
			key, value, err := env.GetKeyValuePair()
			if err != nil {
				return err
			}

			opts, err := env.GetOptions()
			if err != nil {
				return err
			}

			envString := "- " + key + ": " + value
			log.Debugln(envString)
			if !*opts.IsExpand {
				expandString := "  " + "isExpand" + ": " + "false"
				log.Debugln(expandString)
			}
		}
	}

	return nil
}

// readMax reads maximum the given amount of bytes
func readMax(reader io.Reader, maxSize uint) (string, error) {
	p := make([]byte, maxSize)
	if _, err := reader.Read(p); err != nil && err != io.EOF {
		return "", err
	}
	var r []byte
	for _, b := range p {
		if b != 0 {
			r = append(r, b)
		}
	}
	return string(r), nil
}

// readMaxWithTimeout reads maximum a given amount of time using readMax
func readMaxWithTimeout(reader io.Reader, maxSize uint, timeout time.Duration) (value string, err error) {
	valueChan := make(chan string)
	errChan := make(chan error)

	go func() {
		v, err := readMax(reader, maxSize)
		if err != nil {
			errChan <- err
			return
		}
		valueChan <- v
	}()

	select {
	case value = <-valueChan:
	case err = <-errChan:
	case <-time.After(timeout):
		err = errTimeout
	}
	return
}

func add(c *cli.Context) error {
	log.Debugln("[ENVMAN] Work path:", envman.CurrentEnvStoreFilePath)

	key := c.String(KeyKey)
	expand := !c.Bool(NoExpandKey)
	replace := !c.Bool(AppendKey)
	skipIfEmpty := c.Bool(SkipIfEmptyKey)

	var value string

	// read flag value
	if c.IsSet(ValueKey) {
		value = c.String(ValueKey)
		log.Debugf("adding flag value: (%s)", value)
	}

	// read flag file
	if value == "" && c.String(ValueFileKey) != "" {
		var err error
		if value, err = loadValueFromFile(c.String(ValueFileKey)); err != nil {
			log.Fatal("[ENVMAN] Failed to read file value: ", err)
		}
		log.Debugf("adding file flag value: (%s)", value)
	}

	// read piped stdin value
	if value == "" {
		info, err := os.Stdin.Stat()
		if err != nil {
			log.Fatalf("[ENVMAN] Failed to get standard input FileInfo: %s", err)
		}
		if info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
			log.Debugf("adding from piped stdin")

			configs, err := envman.GetConfigs()
			if err != nil {
				log.Fatalf("[ENVMAN] Failed to load envman config: %s", err)
			}

			value, err = readMaxWithTimeout(os.Stdin, uint(configs.EnvBytesLimitInKB*1024), 1*time.Second)
			if err != nil {
				if err == errTimeout {
					log.Warning("[ENVMAN] Standard input read timed out")
				} else {
					log.Fatalf("[ENVMAN] Failed to read standard input: %s", err)
				}
			}

			log.Debugf("stdin value: (%s)", value)
		}
	}

	if err := addEnv(key, value, expand, replace, skipIfEmpty); err != nil {
		log.Fatal("[ENVMAN] Failed to add env:", err)
	}

	log.Debugln("[ENVMAN] Env added")

	if err := logEnvs(); err != nil {
		log.Fatal("[ENVMAN] Failed to print:", err)
	}

	return nil
}
