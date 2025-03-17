// Group all windows in the current workspace, or ungroup, basically similar to
// how i3/sway tabbed container works.
// This script works better with "master" layouts (since the layout is more
// predictable), but it also works in "dwindle" layouts as long the layout is
// not too "deep" (e.g.: too many windows in the same workspace). See
// https://github.com/hyprwm/Hyprland/issues/2822 for more details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thiagokokada/hyprland-go"
)

func must1[T any](v T, err error) T {
	must(err)
	return v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var tabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "Group or ungroup all windows in the current workspace",
	Long:  `Group all windows in the current workspace, or ungroup, similar to how i3/sway tabbed container works.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := hyprland.MustClient()

		aWindow := must1(client.ActiveWindow())
		if len(aWindow.Grouped) > 0 {
			must1(client.Dispatch(
				// If we are already in a group, ungroup
				"togglegroup",
				// Make the current window as master (when using master layout)
				"layoutmsg swapwithmaster master",
			))
		} else {
			var cmdbuf []string
			aWorkspace := must1(client.ActiveWorkspace())
			clients := must1(client.Clients())

			// Grab all windows in the active workspace
			var windows []string
			for _, c := range clients {
				if c.Workspace.Id == aWorkspace.Id {
					windows = append(windows, c.Address)
				}
			}

			// Start by creating a new group
			cmdbuf = append(cmdbuf, "togglegroup")
			for _, w := range windows {
				// Move each window inside the group
				// Once is not enough in case of very "deep" layouts,
				// so we run this multiple times to try to make sure it
				// will work
				// For master layouts we also call swapwithmaster, this
				// makes the switch more reliable
				// TODO: this workaround could be fixed if hyprland
				// supported moving windows based on address and not
				// only positions
				for i := 0; i < 2; i++ {
					cmdbuf = append(cmdbuf, fmt.Sprintf("focuswindow address:%s", w))
					cmdbuf = append(cmdbuf, "layoutmsg swapwithmaster auto")
					cmdbuf = append(cmdbuf, "moveintogroup l")
					cmdbuf = append(cmdbuf, "moveintogroup r")
					cmdbuf = append(cmdbuf, "moveintogroup u")
					cmdbuf = append(cmdbuf, "moveintogroup d")
				}
			}
			// Focus in the active window at the end
			cmdbuf = append(cmdbuf, fmt.Sprintf("focuswindow address:%s", aWindow.Address))

			// Dispatch buffered commands in one call for performance,
			// hyprland-go will take care of splitting it in smaller calls
			// if necessary
			must1(client.Dispatch(cmdbuf...))
		}
	},
}

func init() {
	rootCmd.AddCommand(tabsCmd)
}
