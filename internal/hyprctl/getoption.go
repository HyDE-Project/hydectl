// TODO: See https://github.com/thiagokokada/hyprland-go/issues/44
// // As of the moment hyprland-go do not handle Float and only handles Int.
// This is a limitation of the library and the only way to handle Float is to use the hyprctl package directly.
package hyprctl

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Option struct {
	Option string  `json:"option"`
	Int    int     `json:"int,omitempty"`
	Float  float64 `json:"float,omitempty"`
	Set    bool    `json:"set"`
}

func GetOption(optionName string) (*Option, error) {
	cmd := exec.Command("hyprctl", "getoption", "-j", optionName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute hyprctl: %w", err)
	}

	var option Option
	if err := json.Unmarshal(output, &option); err != nil {
		return nil, fmt.Errorf("failed to unmarshal option: %w", err)
	}

	return &option, nil
}
