package cmd

import (
	"fmt"
	"os/exec"

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
			fmt.Println("Reloading HyDE configuration...")
			out, err := exec.Command("hyde-shell", "reload").Output()
			if err != nil {
				fmt.Printf("Error executing command: %v\n", err)
				return
			}
			fmt.Printf("Command output: %s\n", out)
			fmt.Println("Configuration reloaded successfully.")
		},
	},
}

// Register built-in commands
func init() {
	for _, cmd := range builtins {
		rootCmd.AddCommand(cmd)
	}
}
