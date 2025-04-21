package cmd

import (
	"fmt"
	"time"

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
	step      float64
)

var zoomCmd = &cobra.Command{
	Use:   "zoom",
	Short: "Zoom in/out Hyprland",
	Long:  `Zoom in/out Hyprland or reset the zoom level.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !zoomIn && !zoomOut && !zoomReset {
			fmt.Println("Usage: zoom --in|--out|--reset [--intensity value] [--step value]")
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

			targetZoomFactor := zoomFactor.Float + intensity

			if step > 0 {
				// Gradual zooming with steps
				currentZoom := zoomFactor.Float
				for currentZoom < targetZoomFactor {
					currentZoom += step
					if currentZoom > targetZoomFactor {
						currentZoom = targetZoomFactor
					}
					_, err = client.Keyword(fmt.Sprintf("cursor:zoom_factor %f", currentZoom))
					if err != nil {
						logger.Errorf("Error setting zoom factor: %v", err)
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
			} else {
				// Immediate zoom
				_, err = client.Keyword(fmt.Sprintf("cursor:zoom_factor %f", targetZoomFactor))
				if err != nil {
					logger.Errorf("Error setting zoom factor: %v", err)
				}
			}
		} else if zoomOut {
			_, err := client.Keyword("cursor:no_hardware_cursors 1")
			if err != nil {
				logger.Errorf("Error setting cursor state: %v", err)
				return
			}

			targetZoomFactor := zoomFactor.Float - intensity
			if targetZoomFactor < 1 {
				targetZoomFactor = 1
			}

			if step > 0 {
				// Gradual zooming out with steps
				currentZoom := zoomFactor.Float
				for currentZoom > targetZoomFactor {
					currentZoom -= step
					if currentZoom < targetZoomFactor {
						currentZoom = targetZoomFactor
					}
					_, err = client.Keyword(fmt.Sprintf("cursor:zoom_factor %f", currentZoom))
					if err != nil {
						logger.Errorf("Error setting zoom factor: %v", err)
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
			} else {
				// Immediate zoom out
				_, err = client.Keyword(fmt.Sprintf("cursor:zoom_factor %f", targetZoomFactor))
				if err != nil {
					logger.Errorf("Error setting zoom factor: %v", err)
				}
			}
		} else if zoomReset {
			if step > 0 && zoomFactor.Float > 1 {
				// Gradual reset with steps
				currentZoom := zoomFactor.Float
				for currentZoom > 1 {
					currentZoom -= step
					if currentZoom < 1 {
						currentZoom = 1
					}
					_, err = client.Keyword(fmt.Sprintf("cursor:zoom_factor %f", currentZoom))
					if err != nil {
						logger.Errorf("Error resetting zoom factor: %v", err)
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
			} else {
				// Immediate reset
				_, err := client.Keyword("cursor:zoom_factor 1")
				if err != nil {
					logger.Errorf("Error resetting zoom factor: %v", err)
				}
			}
		}
	},
}

func init() {
	zoomCmd.Flags().BoolVarP(&zoomIn, "in", "i", false, "Zoom in")
	zoomCmd.Flags().BoolVarP(&zoomOut, "out", "o", false, "Zoom out")
	zoomCmd.Flags().BoolVarP(&zoomReset, "reset", "r", false, "Reset zoom")
	zoomCmd.Flags().Float64Var(&intensity, "intensity", 0.1, "Zoom intensity")
	zoomCmd.Flags().Float64VarP(&step, "step", "s", 0, "Granular step for gradual zooming (0 for immediate zoom)")
	rootCmd.AddCommand(zoomCmd)
}
