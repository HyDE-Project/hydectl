package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"hydectl/internal/logging"
)

// LoadScripts searches for executable scripts in the specified directories.
func LoadScripts(dirs []string) ([]string, error) {
	var scripts []string
	scriptMap := make(map[string]bool)

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			logging.Debugf("Directory does not exist: %s", dir)
			continue // Skip if the directory does not exist
		}
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || strings.HasPrefix(info.Name(), ".") || info.Mode().Perm()&0111 == 0 {
				return nil // Skip directories, hidden files, and non-executable files
			}
			baseName := strings.Split(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), ".")[0]
			if !scriptMap[baseName] {
				scripts = append(scripts, baseName)
				scriptMap[baseName] = true
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return scripts, nil
}

// ExecuteScript runs the specified script with the provided arguments.
func ExecuteScript(script string, args []string) error {
	var cmd *exec.Cmd
	switch filepath.Ext(script) {
	case ".sh":
		cmd = exec.Command("bash", append([]string{script}, args...)...)
	case ".py":
		cmd = exec.Command("python", append([]string{script}, args...)...)
	default:
		cmd = exec.Command(script, args...)
	}

	logging.Infof("Executing script: %s with args: %v", script, args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logging.Errorf("Failed to execute script %s: %v", script, err)
		return fmt.Errorf("failed to execute script %s: %w", script, err)
	}

	return nil
}

// ValidateScript checks if the script exists and is executable.
func ValidateScript(script string) (bool, error) {
	info, err := os.Stat(script)
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
