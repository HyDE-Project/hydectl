package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hydectl",
	Short: "Hydectl is a CLI tool for managing HyDE configurations and scripts.",
	Long:  `Hydectl provides a command-line interface for executing built-in commands and user-defined scripts.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you can define flags and configuration settings.
	// For example: rootCmd.PersistentFlags().StringVar(&someVar, "someFlag", "", "Description of someFlag")
}
