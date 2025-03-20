package cmd

import (
	"fmt"

	"hydectl/internal/hydeshell"

	"github.com/spf13/cobra"
)

var (
	skipClone    bool
	fetchTheme   string
	selectTheme  bool
	listThemes   bool
	jsonOutput   bool
	previewTheme string
	previewText  string
	themeName    string
	themeURL     string
)

// themeCmd represents the base command for theme operations
var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage themes",
	Long:  `Manage themes including selecting, navigating, importing, and listing available themes.`,
}

// themeSelectCmd represents the "theme select" command
var themeSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select a theme",
	Run: func(cmd *cobra.Command, args []string) {
		hydeshell.RunCommand("theme.select")
	},
}

// themeNextCmd represents the "theme next" command
var themeNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Switch to the next theme",
	Run: func(cmd *cobra.Command, args []string) {
		hydeshell.RunCommand("theme.switch", "-n")
	},
}

// themePrevCmd represents the "theme prev" command
var themePrevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Switch to the previous theme",
	Run: func(cmd *cobra.Command, args []string) {
		hydeshell.RunCommand("theme.switch", "-p")
	},
}

// themeSetCmd represents the "theme set" command
var themeSetCmd = &cobra.Command{
	Use:   "set [theme name]",
	Short: "Set a theme using the specified theme name",
	Long:  "Set a theme using the specified theme name.",
	Args:  cobra.ExactArgs(1), // Ensure exactly one argument is provided
	Run: func(cmd *cobra.Command, args []string) {
		themeName := args[0]
		hydeshell.RunCommand("theme.switch", "-s", themeName)
	},
}

// themeImportCmd represents the "theme import" command
var themeImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import themes",
	Long:  `Imports themes from the hyde-gallery repository or a specific theme by name and URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		if themeName != "" && themeURL != "" {
			fmt.Printf("Importing theme: %s from URL: %s\n", themeName, themeURL)
			hydeshell.RunCommand("theme.patch", themeName, themeURL)
			return
		}

		if jsonOutput {
			fmt.Println("Fetching JSON data after cloning.")
			hydeshell.RunCommand("theme.import", "--json")
		}
		if previewTheme != "" {
			hydeshell.RunCommand("theme.import", "--preview", previewTheme)
		}
		if previewText != "" {
			hydeshell.RunCommand("theme.import", "--preview-text", previewText)
		}
		if skipClone {
			fmt.Println("Importing themes with --skip-clone.")
			hydeshell.RunCommand("theme.import", "--skip-clone")
		}
		if fetchTheme != "" {
			hydeshell.RunCommand("theme.import", "--fetch", fetchTheme)
		}
		if selectTheme {
			hydeshell.RunCommand("theme.import", "--select")
		}
		if listThemes {
			hydeshell.RunCommand("theme.import", "--list")
		}
		if !jsonOutput && previewTheme == "" && previewText == "" && !skipClone && fetchTheme == "" && !selectTheme && !listThemes {
			hydeshell.RunCommand("theme.import", "--select")
		}
	},
}

func init() {
	// Add subcommands to themeCmd
	themeCmd.AddCommand(themeSelectCmd)
	themeCmd.AddCommand(themeNextCmd)
	themeCmd.AddCommand(themePrevCmd)
	themeCmd.AddCommand(themeSetCmd)

	// Add flags to themeImportCmd
	themeImportCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Fetch JSON data after cloning")
	themeImportCmd.Flags().BoolVar(&skipClone, "skip-clone", false, "Skip cloning repository")
	themeImportCmd.Flags().StringVarP(&fetchTheme, "fetch", "f", "", "Fetch and update a specific theme by name ('all' to fetch all themes)")
	themeImportCmd.Flags().StringVar(&themeName, "name", "", "Name of the theme to import")
	themeImportCmd.Flags().StringVar(&themeURL, "url", "", "URL of the theme to import. Accepts a URL or a path to a local theme")
	themeImportCmd.MarkFlagsRequiredTogether("name", "url")

	themeCmd.AddCommand(themeImportCmd)
	rootCmd.AddCommand(themeCmd)
}
