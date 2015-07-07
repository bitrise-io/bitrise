package cli

import "github.com/codegangsta/cli"

func init() {
	cli.AppHelpTemplate = `NAME: {{.Name}} - {{.Usage}}

USAGE: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

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
COMMAND HELP: {{.Name}} COMMAND --help/-h

`
}
