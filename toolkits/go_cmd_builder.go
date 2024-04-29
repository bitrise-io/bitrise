package toolkits

import "github.com/bitrise-io/go-utils/command"

type goCmdBuilder struct {
	goConfig GoConfigurationModel
}

func (g goCmdBuilder) goBuildArgs(outputBin string, explicitlyVendor bool) []string {
	// -mod=vendor is used to explicitly enable vendoring. (Go 1.14 and later defaults to this if vendor dir is present.)
	// When vendoring is enabled, build commands like go build and go test load packages
	// from the vendor directory instead of accessing the network or the local module cache.
	args := []string{"build"}
	if explicitlyVendor {
		args = append(args, "-mod=vendor")
	}
	args = append(args, "-ldflags", "-w")

	args = append(args, "-o", outputBin)

	return args
}

func (g goCmdBuilder) goBuildEnv(stepAbsDir string, shouldCheckGoSum bool) []string {
	// GOSUMDB=off disables requests to checksum databases (https://golang.org/ref/mod#communicating-with-proxies)
	// There is no go.sum file as migrated go modules are in use,
	// so it is used only for an extra measure disabling network queries.
	envs := []string{
		"PWD=" + stepAbsDir,
		"GOROOT=" + g.goConfig.GOROOT,
		"GO111MODULE=on",
	}
	if !shouldCheckGoSum {
		envs = append(envs, "GOSUMDB=off")
	}

	return envs
}

func (g goCmdBuilder) goBuildMigratedModules(stepAbsDir string, outputBin string) *command.Model {
	buildCmd := command.New(g.goConfig.GoBinaryPath, g.goBuildArgs(outputBin, true)...).
		AppendEnvs(g.goBuildEnv(stepAbsDir, false)...).
		SetDir(stepAbsDir)

	return buildCmd
}

func (g goCmdBuilder) goBuildInModuleMode(stepAbsDir string, outputBin string) *command.Model {
	buildCmd := command.New(g.goConfig.GoBinaryPath, g.goBuildArgs(outputBin, false)...).
		AppendEnvs(g.goBuildEnv(stepAbsDir, true)...).
		SetDir(stepAbsDir)

	return buildCmd
}
