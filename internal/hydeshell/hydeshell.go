package hydeshell

import (
	"fmt"
	"os/exec"
)

// RunCommand executes the hyde-shell command with the given arguments.
func RunCommand(command string, args ...string) error {
	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("hyde-shell", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute hyde-shell command: %w\nOutput: %s", err, string(output))
	}
	fmt.Println(string(output))
	return nil
}
