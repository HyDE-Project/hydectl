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
	Short: "hydectl is a CLI tool for managing HyDE configurations and scripts.",
	Long:  `hydectl provides a command-line interface for executing built-in commands and user-defined scripts.`,
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
}
