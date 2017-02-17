package cli

import (
	"log"
	"strings"

	"fmt"

	"github.com/bitrise-io/bitrise/plugins"
	"github.com/urfave/cli"
)

const (
	helpTemplate = `
NAME: {{.Name}} - {{.Usage}}

USAGE: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND/PLUGIN [arg...]

VERSION: {{.Version}}{{if or .Author .Email}}

AUTHOR:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
GLOBAL OPTIONS:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
COMMANDS:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
%s
COMMAND HELP: {{.Name}} COMMAND --help/-h

`
)

func getPluginsList() string {
	plugins.InitPaths()

	pluginListString := "PLUGINS:\n"

	pluginList, err := plugins.InstalledPluginList()
	if err != nil {
		log.Fatalf("Failed to list plugins, error: %s", err)
	}

	if len(pluginList) > 0 {
		plugins.SortByName(pluginList)
		for _, plugin := range pluginList {
			pluginListString += fmt.Sprintf("  :%s\t%s\n", plugin.Name, strings.Split(plugin.Description, "\n")[0])
		}
	} else {
		pluginListString += "  No plugins installed\n"
	}

	return pluginListString
}

func initAppHelpTemplate() {
	cli.AppHelpTemplate = fmt.Sprintf(helpTemplate, getPluginsList())
}
