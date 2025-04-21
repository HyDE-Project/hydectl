package main

import (
	"os"
	"path/filepath"

	"hydectl/cmd"
	"hydectl/internal/logger"
	"hydectl/internal/plugin"

	"github.com/adrg/xdg"
)

func main() {
	logger.SetupLogging()

	cmd.ScriptPaths = []string{
		xdg.ConfigHome + "/lib/hydectl/scripts",
		// os.Getenv("HOME") + "/.local/lib/hyde",
		os.Getenv("HOME") + "/.local/lib/hydectl/scripts",
		"/usr/local/lib/hydectl/scripts",
		"/usr/lib/hydectl/scripts",
	}

	if len(os.Args) > 1 {
		scriptName := os.Args[1]
		var scriptPath string
		for _, dir := range cmd.ScriptPaths {
			path := filepath.Join(dir, scriptName)
			if _, err := os.Stat(path); err == nil {
				scriptPath = path
				break
			}
		}

		if scriptPath == "" {
			// Try to find the script with a known extension
			for _, dir := range cmd.ScriptPaths {
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
			cmd.Execute()
			return
		}

		logger.Debugf("Executing script: %s", scriptPath)
		plugin.ExecuteScript(scriptPath, os.Args[2:])
		return
	}

	cmd.Execute()
}
