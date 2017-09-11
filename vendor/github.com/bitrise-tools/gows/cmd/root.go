package cmd

import (
	"errors"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-tools/gows/config"
	"gopkg.in/viktorbenei/cobra.v0"
)

var (
	loglevelFlag string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gows",
	Short: "Go Workspace / Environment Manager, to easily manage the Go Workspace during development.",
	Long: `Go Workspace / Environment Manager, to easily manage the Go Workspace during development.

Work in isolated (development) environment when you're working on your Go projects.
No cross-project dependency version missmatch, no more packages left out from vendor/.

No need for initializing a go workspace either, your project can be located anywhere,
not just in a predefined $GOPATH workspace. gows will take care about crearing
the (per-project isolated) workspace directory structure, no matter where your project is located.

gows works perfectly with other Go tools, all it does is it ensures that every project
gets it's own, isolated Go workspace and sets $GOPATH accordingly.

Sync Mode can be set in the .gows.user.yml config file,
or through the $GOWS_SYNC_MODE environment variable.`,

	DisableFlagParsing: true,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		initLogFormatter()

		// Log level
		if loglevelFlag == "" {
			if loglevelEnv := os.Getenv("GOWS_LOGLEVEL"); loglevelEnv != "" {
				loglevelFlag = loglevelEnv
			} else {
				// default
				loglevelFlag = "info"
			}
		}

		level, err := log.ParseLevel(loglevelFlag)
		if err != nil {
			return err
		}
		log.SetLevel(level)

		return nil
	},
}

func initLogFormatter() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "15:04:05",
	})
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&loglevelFlag, "loglevel", "l", "", `Log level (options: debug, info, warn, error, fatal, panic). [$GOWS_LOGLEVEL]`)
	RootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("No command specified!")
		}
		RootCmd.SilenceErrors = true
		RootCmd.SilenceUsage = true
		return nil
	}
	RootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		cmdName := args[0]
		if cmdName == "-h" || cmdName == "--help" {
			if err := RootCmd.Help(); err != nil {
				return err
			}
			return nil
		}

		cmdArgs := []string{}
		if len(args) > 1 {
			cmdArgs = args[1:]
		}

		userConfig, err := config.LoadUserConfigFromFile()
		if err != nil {
			log.Debug("No User Config found, using defaults")
			userConfig = config.CreateDefaultUserConfig()
		}
		forceSyncMode := os.Getenv("GOWS_SYNC_MODE")
		if forceSyncMode != "" {
			log.Debugf(" (i) Sync Mode specified as a parameter, using it (%s)", forceSyncMode)
			userConfig.SyncMode = forceSyncMode
		}
		log.Debugf("User Config: %#v", userConfig)

		exitCode, err := PrepareEnvironmentAndRunCommand(userConfig, cmdName, cmdArgs...)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		if err != nil {
			return fmt.Errorf("Exit Code was 0, but an error happened: %s", err)
		}
		return nil
	}
}
