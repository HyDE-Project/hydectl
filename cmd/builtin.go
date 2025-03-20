package cmd

import (
	"fmt"
	"hydectl/internal/hydeshell"

	"github.com/spf13/cobra"
)

// Built-in commands
var builtins = []*cobra.Command{
	{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hydectl %s\n", Version)
		},
	},
	{
		Use:   "reload",
		Short: "Reload the HyDE configuration",
		Run: func(cmd *cobra.Command, args []string) {
			hydeshell.RunCommand("reload")
		},
	},
}

// Register built-in commands
func init() {
	for _, cmd := range builtins {
		rootCmd.AddCommand(cmd)
	}
}
