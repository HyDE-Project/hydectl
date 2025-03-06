package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LoadScripts searches for executable scripts in the specified directory.
func LoadScripts(dir string) ([]string, error) {
	var scripts []string

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Mode().Perm()&0111 != 0 { // Check if the file is executable
			scripts = append(scripts, file.Name())
		}
	}

	return scripts, nil
}

// ExecuteScript runs the specified script with the provided arguments.
func ExecuteScript(script string, args []string) error {
	scriptPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "hyprctl", "scripts", script)

	cmd := exec.Command(scriptPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute script %s: %w", script, err)
	}

	return nil
}

// ValidateScript checks if the script exists and is executable.
func ValidateScript(script string) (bool, error) {
	scriptPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "hyprctl", "scripts", script)
	info, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.Mode().Perm()&0111 != 0, nil
}

// GetHelpMessage generates a help message based on available scripts.
func GetHelpMessage(scripts []string) string {
	var sb strings.Builder
	sb.WriteString("Available Scripts:\n")
	for _, script := range scripts {
		sb.WriteString(fmt.Sprintf("- %s\n", script))
	}
	return sb.String()
}
