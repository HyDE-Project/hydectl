// ? CONTRIB NOTE: This is just a CLI handler, the actual implementation is wallpaper.sh
// TODO: If you want to reimplement a full go implementation, Feel free to open a MR.

package cmd

import (
	"hydectl/internal/hydeshell"
	"hydectl/internal/logger"

	"github.com/spf13/cobra"
)

var (
	wallpaperSources []string
	skipStrays       bool
	verbose          bool
	wallpaperBackend string
	wallpaperPath    string
	wallpaperOutput  string
	setAsGlobal      bool
)

var wallpaperCmd = &cobra.Command{
	Use:   "wallpaper",
	Short: "Manage wallpapers",
	Long:  `Manage wallpapers`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List wallpapers",
	Long:  `List wallpapers in the specified directories.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("wallpaper", "--json")
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

var wallSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select wallpaper using rofi",
	Long:  `Select wallpaper using rofi.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("wallpaper", "--select")
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Set next wallpaper",
	Long:  `Set next wallpaper.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("wallpaper", "--next")
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

var previousCmd = &cobra.Command{
	Use:   "previous",
	Short: "Set previous wallpaper",
	Long:  `Set previous wallpaper.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("wallpaper", "--previous")
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Set random wallpaper",
	Long:  `Set random wallpaper.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("wallpaper", "--random")
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

var setCmd = &cobra.Command{
	Use:   "set [wallpaper path]",
	Short: "Set specified wallpaper",
	Long:  `Set specified wallpaper.`,
	Args:  cobra.ExactArgs(1), // Ensure exactly one argument is passed
	Run: func(cmd *cobra.Command, args []string) {
		wallpaperPath = args[0] // Get the wallpaper path from the positional argument

		logger.Debugf("Setting wallpaper to: %s", wallpaperPath)

		err := hydeshell.RunCommand("wallpaper", "--set", wallpaperPath)
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current wallpaper of specified backend",
	Long:  `Get current wallpaper of specified backend.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("wallpaper", "--get")
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

var outputCmd = &cobra.Command{
	Use:   "output",
	Short: "Copy current wallpaper to specified file",
	Long:  `Copy current wallpaper to specified file.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hydeshell.RunCommand("wallpaper", "--output", wallpaperOutput)
		if err != nil {
			logger.Errorf("Error executing hyde-shell command: %v", err)
		}
	},
}

func init() {
	listCmd.Flags().StringSliceVar(&wallpaperSources, "sources", []string{}, "Directories to search for wallpapers")
	listCmd.Flags().BoolVar(&skipStrays, "skip-strays", false, "Skip stray files")
	listCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	listCmd.Flags().Bool("json", true, "Output in JSON format")

	wallpaperCmd.PersistentFlags().StringVar(&wallpaperBackend, "backend", "swww", "Set wallpaper backend to use (swww, mpvpaper, etc.)")
	wallpaperCmd.PersistentFlags().BoolVar(&setAsGlobal, "global", false, "Set wallpaper as global")

	wallpaperCmd.AddCommand(listCmd)
	wallpaperCmd.AddCommand(wallSelectCmd)
	wallpaperCmd.AddCommand(nextCmd)
	wallpaperCmd.AddCommand(previousCmd)
	wallpaperCmd.AddCommand(randomCmd)
	wallpaperCmd.AddCommand(setCmd)
	wallpaperCmd.AddCommand(getCmd)
	wallpaperCmd.AddCommand(outputCmd)
	rootCmd.AddCommand(wallpaperCmd)
}
