package hydeshell

import (
	"fmt"
	"hydectl/internal/logger"
	"os"
	"os/exec"
)

// RunCommand executes the hyde-shell command with the given arguments.
func RunCommand(command string, args ...string) error {
	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("hyde-shell", cmdArgs...)

	// Set up pipes for real-time output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute hyde-shell command: %w", err)
	}
	return nil
}

// RunCommandSilent executes the hyde-shell command with the given arguments without output.
func RunCommandSilent(command string, args ...string) error {
	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("hyde-shell", cmdArgs...)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute hyde-shell command: %w", err)
	}

	// Log output to debug
	logger.Debugf(string(output))

	return nil
}
