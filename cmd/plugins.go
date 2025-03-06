package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Execute commands from the ~/.local/share/hyprctl/scripts/ directory",
	Long:  "This command allows you to manage and execute external scripts found in ~/.local/share/hyprctl/scripts/",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		scriptName := args[0]
		scriptPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "hyprctl", "scripts", scriptName)

		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Printf("Script %s does not exist.\n", scriptName)
			return
		}

		// Execute the script (this is a placeholder for actual execution logic)
		fmt.Printf("Executing script: %s\n", scriptPath)
		// Here you would add the logic to execute the script
	},
}

func init() {
	rootCmd.AddCommand(pluginsCmd)
}
