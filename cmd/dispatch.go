package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"hydectl/internal/logger"
	"hydectl/internal/plugin"

	"github.com/spf13/cobra"
)

var (
	listPlugins bool
	ScriptPaths []string
)

var dispatchCmd = &cobra.Command{
	Use:   "dispatch [plugin] [args...]",
	Short: "Dispatch a plugin command",
	Long:  `Dispatch a plugin command by specifying the plugin name and arguments.`,
	Run: func(cmd *cobra.Command, args []string) {
		if listPlugins {
			scripts, err := plugin.LoadScripts(ScriptPaths)
			if err != nil {
				logger.Errorf("Error loading scripts: %v", err)
				fmt.Printf("Error loading scripts: %v\n", err)
				return
			}

			fmt.Println("Available Plugins:")
			for _, script := range scripts {
				fmt.Println(script)
			}
			return
		}

		if len(args) < 1 {
			cmd.Help()
			return
		}

		pluginName := args[0]
		pluginArgs := args[1:]

		// Filter out non-existent directories
		existingScriptPaths := filterExistingPaths(ScriptPaths)

		var scriptPath string
		for _, dir := range existingScriptPaths {
			path := filepath.Join(dir, pluginName)
			if _, err := os.Stat(path); err == nil {
				scriptPath = path
				break
			}
		}

		if scriptPath == "" {
			// Try to find the script with a known extension
			for _, dir := range existingScriptPaths {
				for _, ext := range []string{".sh", ".py"} {
					path := filepath.Join(dir, pluginName+ext)
					if _, err := os.Stat(path); err == nil {
						scriptPath = path
						break
					}
				}
				if scriptPath != "" {
					break
				}
			}
		}

		if scriptPath == "" {
			logger.Infof("Plugin %s does not exist.", pluginName)
			fmt.Printf("Plugin %s does not exist.\n", pluginName)
			return
		}

		if err := plugin.ExecuteScript(scriptPath, pluginArgs); err != nil {
			logger.Errorf("Error executing plugin: %v", err)
			fmt.Printf("Error executing plugin: %v\n", err)
		}
	},
}

func filterExistingPaths(paths []string) []string {
	var existingPaths []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		} else {
			logger.Debugf("Directory does not exist: %s", path)
		}
	}
	return existingPaths
}

func init() {
	dispatchCmd.Flags().BoolVarP(&listPlugins, "list", "l", false, "List all available plugins")
	rootCmd.AddCommand(dispatchCmd)
}
