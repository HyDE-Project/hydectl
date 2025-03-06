package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"hydectl/internal/logging"
	"hydectl/internal/plugin"

	"github.com/spf13/cobra"
)

var (
	ScriptPaths []string
	listScripts bool
	Version     = "v0.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "hydectl",
	Short: "hydectl is a CLI tool for managing HyDE configurations and scripts.",
	Long:  `hydectl provides a command-line interface for executing built-in commands and user-defined scripts.`,
	Run: func(cmd *cobra.Command, args []string) {
		if listScripts {
			scripts, err := plugin.LoadScripts(ScriptPaths)
			if err != nil {
				logging.Errorf("Error loading scripts: %v", err)
				fmt.Printf("Error loading scripts: %v\n", err)
				return
			}

			fmt.Println("Available Scripts:")
			for _, script := range scripts {
				fmt.Println(script)
			}
			return
		}

		if len(args) == 0 {
			cmd.Help()
			fmt.Println("\nPlugin Commands:")
			scripts, err := plugin.LoadScripts(ScriptPaths)
			if err != nil {
				logging.Errorf("Error loading scripts: %v", err)
				fmt.Printf("Error loading scripts: %v\n", err)
				return
			}
			const colWidth = 30
			const maxCols = 5
			numScripts := len(scripts)
			numRows := (numScripts + maxCols - 1) / maxCols

			for row := 0; row < numRows; row++ {
				for col := 0; col < maxCols; col++ {
					idx := col*numRows + row
					if idx < numScripts {
						fmt.Printf("%-*s", colWidth, scripts[idx])
					}
				}
				fmt.Println()
			}
			return
		}

		scriptName := args[0]
		var scriptPath string
		for _, dir := range ScriptPaths {
			path := filepath.Join(dir, scriptName)
			if _, err := os.Stat(path); err == nil {
				scriptPath = path
				break
			}
		}

		if scriptPath == "" {
			// Try to find the script with a known extension
			for _, dir := range ScriptPaths {
				for _, ext := range []string{".sh", ".py"} {
					path := filepath.Join(dir, scriptName+ext)
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
			logging.Infof("Script %s does not exist.", scriptName)
			fmt.Printf("Script %s does not exist.\n", scriptName)
			return
		}

		if err := plugin.ExecuteScript(scriptPath, args[1:]); err != nil {
			logging.Errorf("Error executing script: %v", err)
			fmt.Printf("Error executing script: %v\n", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&listScripts, "list", "l", false, "List all available scripts")

	// Add script completion
	rootCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		scripts, err := plugin.LoadScripts(ScriptPaths)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return scripts, cobra.ShellCompDirectiveNoFileComp
	}
}
