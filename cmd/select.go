// This is use to call selectors available hyde.
// Example emoji picker,glyph picker, color picker, etc.
package cmd

import (
	"hydectl/internal/hydeshell"
	"hydectl/internal/logger"

	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select various items",
	Long:  `Select various items such as emoji or glyph.`,
}

var emojiCmd = &cobra.Command{
	Use:   "emoji",
	Short: "Select an emoji",
	Long:  `Select an emoji using rofi.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("emoji-picker", args...)
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

var glyphCmd = &cobra.Command{
	Use:   "glyph",
	Short: "Select a glyph",
	Long:  `Select a glyph using rofi.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("glyph-picker", args...)
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

func init() {
	selectCmd.AddCommand(emojiCmd)
	selectCmd.AddCommand(glyphCmd)
	rootCmd.AddCommand(selectCmd)
}
