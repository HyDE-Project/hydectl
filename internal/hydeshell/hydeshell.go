package hydeshell

import (
	"fmt"
	"hydectl/internal/logger"
	"os"
	"os/exec"
)

func RunCommand(command string, args ...string) error {
	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("hyde-shell", cmdArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute hyde-shell command: %w", err)
	}
	return nil
}

func RunCommandSilent(command string, args ...string) error {
	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("hyde-shell", cmdArgs...)

	cmd.Stdin = os.Stdin

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute hyde-shell command: %w", err)
	}

	logger.Debugf(string(output))

	return nil
}
