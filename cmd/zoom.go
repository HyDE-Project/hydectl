package cmd

import (
	"fmt"

	"hydectl/internal/hyprctl"
	"hydectl/internal/logger"

	"github.com/spf13/cobra"
	"github.com/thiagokokada/hyprland-go"
)

var (
	zoomIn    bool
	zoomOut   bool
	zoomReset bool
	intensity float64
)

var zoomCmd = &cobra.Command{
	Use:   "zoom",
	Short: "Zoom in/out Hyprland",
	Long:  `Zoom in/out Hyprland or reset the zoom level.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !zoomIn && !zoomOut && !zoomReset {
			fmt.Println("Usage: zoom --in|--out|--reset [--intensity value]")
			return
		}

		cursorState, err := hyprctl.GetOption("cursor:no_hardware_cursors")
		if err != nil {
			logger.Errorf("Error getting cursor state: %v", err)
			return
		}
		logger.Infof("Current cursor state: %v", cursorState)

		defer func() {
			client := hyprland.MustClient()
			_, err := client.Keyword(fmt.Sprintf("cursor:no_hardware_cursors %d", cursorState.Int))
			if err != nil {
				logger.Errorf("Error resetting cursor state: %v", err)
			}
		}()

		zoomFactor, err := hyprctl.GetOption("cursor:zoom_factor")
		if err != nil {
			logger.Errorf("Error getting zoom factor: %v", err)
			return
		}
		logger.Infof("Current zoom factor: %v", zoomFactor)

		client := hyprland.MustClient()

		if zoomIn {
			_, err := client.Keyword("cursor:no_hardware_cursors 1")
			if err != nil {
				logger.Errorf("Error setting cursor state: %v", err)
				return
			}
			newZoomFactor := zoomFactor.Float + intensity
			_, err = client.Keyword(fmt.Sprintf("cursor:zoom_factor %f", newZoomFactor))
			if err != nil {
				logger.Errorf("Error setting zoom factor: %v", err)
			}
		} else if zoomOut {
			_, err := client.Keyword("cursor:no_hardware_cursors 1")
			if err != nil {
				logger.Errorf("Error setting cursor state: %v", err)
				return
			}
			newZoomFactor := zoomFactor.Float - intensity
			if newZoomFactor < 1 {
				newZoomFactor = 1
			}
			_, err = client.Keyword(fmt.Sprintf("cursor:zoom_factor %f", newZoomFactor))
			if err != nil {
				logger.Errorf("Error setting zoom factor: %v", err)
			}
		} else if zoomReset {
			_, err := client.Keyword("cursor:zoom_factor 1")
			if err != nil {
				logger.Errorf("Error resetting zoom factor: %v", err)
			}
		}
	},
}

func init() {
	zoomCmd.Flags().BoolVar(&zoomIn, "in", false, "Zoom in")
	zoomCmd.Flags().BoolVar(&zoomOut, "out", false, "Zoom out")
	zoomCmd.Flags().BoolVar(&zoomReset, "reset", false, "Reset zoom")
	zoomCmd.Flags().Float64Var(&intensity, "intensity", 0.1, "Zoom intensity")
	rootCmd.AddCommand(zoomCmd)
}
