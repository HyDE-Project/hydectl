package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "v0.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "hydectl",
	Short: "Tool for interacting with HyDE",
	Long:  `HyDE Projects Official Command line interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		fmt.Printf("Unknown command: %s\n", args[0])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add script completion
	rootCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Set custom help template
	rootCmd.SetHelpTemplate(getHelpTemplate())
}

func getHelpTemplate() string {
	const (
		magenta = "\033[35m"
		red     = "\033[31m"
		cyan    = "\033[36m"
		yellow  = "\033[33m"
		green   = "\033[32m"
		blue    = "\033[34m"
		reset   = "\033[0m"
		bold    = "\033[1m"
	)

	return `
    {{with (or .Long .Short)}}` + cyan + `{{.}}` + reset + `{{end}}

` + bold + yellow + `Usage:` + reset + `
  {{.UseLine}}` + reset + `

` + bold + yellow + `Available Commands:` + reset + `
{{range .Commands}}{{if (and .IsAvailableCommand (ne .Name "help"))}}
  ` + green + `{{rpad .Name .NamePadding }}` + reset + ` {{.Short}}{{end}}{{end}}

` + bold + yellow + `Flags:` + reset + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

` + bold + yellow + `Tips:` + reset + `
  ` + blue + `Use "{{.CommandPath}} [command] --help" for more information about a command.` + reset + `
  ` + magenta + `To PASS additional arguments directly to the command, append '--' before the arguments.` + reset + `

` + bold + yellow + `hydectl version: ` + reset + Version + `
`
}
