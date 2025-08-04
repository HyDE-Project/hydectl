package plugin

import (
	"fmt"
	"hydectl/internal/logger"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	scriptCache = make(map[string]string)
	cacheMutex  = &sync.Mutex{}
	cached      = false
)

func FindAllScripts(dirs []string) (map[string]string, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cached {
		return scriptCache, nil
	}

	scriptCache = make(map[string]string)
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			logger.Debugf("Directory does not exist: %s", dir)
			continue // Skip if the directory does not exist
		}
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || strings.HasPrefix(info.Name(), ".") || info.Mode().Perm()&0111 == 0 {
				return nil // Skip directories, hidden files, and non-executable files
			}
			baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			if _, exists := scriptCache[baseName]; !exists {
				scriptCache[baseName] = path
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	cached = true
	return scriptCache, nil
}

func LoadScripts(dirs []string) ([]string, error) {
	scripts, err := FindAllScripts(dirs)
	if err != nil {
		return nil, err
	}

	var scriptNames []string
	for name := range scripts {
		scriptNames = append(scriptNames, name)
	}
	return scriptNames, nil
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

	logger.Infof("Executing script: %s with args: %v", script, args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Errorf("Failed to execute script %s: %v", script, err)
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
