package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
// printRootHelp renders the root help in the layout the previous framework used
// (NAME / USAGE / VERSION / GLOBAL OPTIONS / COMMANDS / PLUGINS), including the
// installed plugin list. Subcommands keep cobra's native help.
func printRootHelp(root *cobra.Command) {
	w := tabwriter.NewWriter(root.OutOrStdout(), 0, 8, 2, ' ', 0)

	fmt.Fprintf(w, "\nNAME: %s - %s\n\n", root.Name(), root.Short)
	fmt.Fprintf(w, "USAGE: %s [OPTIONS] COMMAND/PLUGIN [arg...]\n\n", root.Name())
	fmt.Fprintf(w, "VERSION: %s\n\n", root.Version)

	fmt.Fprintln(w, "GLOBAL OPTIONS:")
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		name := "--" + f.Name
		if f.Shorthand != "" {
			name += ", -" + f.Shorthand
		}
		usage := f.Usage
		if envs := f.Annotations[envVarAnnotation]; len(envs) > 0 {
			usage += " [$" + envs[0] + "]"
		}
		fmt.Fprintf(w, "  %s\t%s\n", name, usage)
	})
	fmt.Fprintf(w, "  %s\t%s\n", "--help, -h", "show help")
	fmt.Fprintf(w, "  %s\t%s\n", "--version, -v", "print the version")

	fmt.Fprintln(w, "\nCOMMANDS:")
	for _, c := range root.Commands() {
		if c.Hidden {
			continue
		}
		name := c.Name()
		if len(c.Aliases) > 0 {
			name += ", " + strings.Join(c.Aliases, ", ")
		}
		fmt.Fprintf(w, "  %s\t%s\n", name, c.Short)
	}

	fmt.Fprintf(w, "\n%s\n", getPluginsList())
	fmt.Fprintf(w, "COMMAND HELP: %s COMMAND --help/-h\n\n", root.Name())

	_ = w.Flush()
}

func getPluginsList() string {
	pluginListString := "PLUGINS:\n"

	pluginList, err := plugins.InstalledPluginList()
	if err != nil {
		failf("Failed to list plugins, error: %s", err)
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
