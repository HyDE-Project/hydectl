//  TODO: Dynamic Completion and Help

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"hydectl/internal/logger"
	"hydectl/internal/plugin"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

var dynamicCommands []*cobra.Command

func init() {
	logger.Debug("Initialize dispatch command")

	// Add dynamic commands
	logger.Debug("Add plugin commands")
	AddPluginCommands()

	dispatchCmd.Flags().BoolVarP(&listPlugins, "list", "l", false, "List all available plugins")
	dispatchCmd.SetHelpFunc(customHelpFunc)
	rootCmd.AddCommand(dispatchCmd)
}

func customHelpFunc(cmd *cobra.Command, args []string) {
	fmt.Println("Custom Help Message for Dispatch Command")
	fmt.Println("Usage:")
	fmt.Printf("  %s\n", cmd.UseLine())
	fmt.Println(cmd.Long)
	fmt.Println("\nAvailable Commands:")
	for _, c := range cmd.Commands() {
		fmt.Printf("  %s\t%s\n", c.Name(), c.Short)
	}
	fmt.Println("\nFlags:")
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		fmt.Printf("  --%s\t%s\n", flag.Name, flag.Usage)
	})
	fmt.Println("\nUse \"dispatch [command] --help\" for more information about a command.")
}

// AddCommand dynamically adds a new command to the CLI
func AddCommand(use, short, long string, run func(cmd *cobra.Command, args []string)) {
	newCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run:   run,
	}
	dynamicCommands = append(dynamicCommands, newCmd)
	rootCmd.AddCommand(newCmd)
	logger.Debugf("Command %s added successfully", use)
}

type ScriptUsage struct {
	Use     string `json:"Use"`
	Short   string `json:"Short"`
	Long    string `json:"Long"`
	Options []struct {
		Name    string `json:"Name"`
		Short   string `json:"Short"`
		Long    string `json:"Long"`
		Type    string `json:"Type"`
		Default bool   `json:"Default"`
	} `json:"Options"`
}

// Example function to dynamically add plugin commands
func AddPluginCommands() {
	logger.Debug("Loading scripts for dynamic commands")
	scripts, err := plugin.LoadScripts(ScriptPaths)
	if err != nil {
		logger.Errorf("Error loading scripts: %v", err)
		return
	}

	for _, script := range scripts {
		logger.Debugf("Processing script: %s", script)
		scriptPath := findScriptPath(script)
		if scriptPath == "" {
			logger.Infof("Plugin %s does not exist.", script)
			continue
		}

		usage, err := getScriptUsage(scriptPath)
		if err != nil {
			logger.Errorf("Error getting usage for script %s: %v", script, err)
			continue
		}

		logger.Debugf("Adding command: %s", usage.Use)
		newCmd := &cobra.Command{
			Use:   usage.Use,
			Short: usage.Short,
			Long:  usage.Long,
			Run: func(cmd *cobra.Command, args []string) {
				logger.Debugf("Executing script: %s with args: %v", scriptPath, args)
				if err := plugin.ExecuteScript(scriptPath, args); err != nil {
					logger.Errorf("Error executing plugin: %v", err)
					fmt.Printf("Error executing plugin: %v\n", err)
				}
			},
		}

		for _, option := range usage.Options {
			logger.Debugf("Adding option: %s", option.Name)
			switch option.Type {
			case "bool":
				newCmd.Flags().Bool(option.Name, option.Default, option.Long)
				// Add other types as needed
			}
		}

		dynamicCommands = append(dynamicCommands, newCmd)
		rootCmd.AddCommand(newCmd)
		logger.Debugf("Command %s added successfully", usage.Use)
	}
}

func getScriptUsage(scriptPath string) (*ScriptUsage, error) {
	logger.Debugf("Getting usage for script: %s", scriptPath)
	cmd := exec.Command(scriptPath, "__usage__")
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("Error executing script for usage: %v", err)
		return nil, err
	}

	logger.Debugf("Script usage output: %s", output)
	var usage ScriptUsage
	if err := json.Unmarshal(output, &usage); err != nil {
		logger.Errorf("Error unmarshalling usage JSON: %v", err)
		return nil, err
	}

	return &usage, nil
}

func findScriptPath(pluginName string) string {
	existingScriptPaths := filterExistingPaths(ScriptPaths)
	for _, dir := range existingScriptPaths {
		path := filepath.Join(dir, pluginName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
		for _, ext := range []string{".sh", ".py"} {
			path := filepath.Join(dir, pluginName+ext)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}
