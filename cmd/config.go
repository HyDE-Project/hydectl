package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"hydectl/internal/config"
	"hydectl/internal/tui"
)

var previewHighlightStyle string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Interactive configuration file editor",
	Long:  `Open an interactive selector to edit application configuration files with pre/post hooks.`,
	Run:   runConfigCommand,
}

func init() {
	configCmd.Flags().StringVar(&previewHighlightStyle, "preview-highlight", "monokai", "Syntax highlight style for preview (e.g. monokai, dracula, solarized-dark, etc)")
	rootCmd.AddCommand(configCmd)
}

func runConfigCommand(cmd *cobra.Command, args []string) {
	registry, err := config.LoadConfigRegistry()
	if err != nil {
		fmt.Printf("Error loading config registry: %v\n", err)
		return
	}

	if len(registry.Apps) == 0 {
		fmt.Println("No applications found in config registry.")
		fmt.Println("Please add applications to your config-registry.toml file.")
		return
	}

	debug, _ := cmd.Flags().GetBool("debug")
	model := tui.NewModel(registry, previewHighlightStyle, debug)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		return
	}

	if m, ok := finalModel.(*tui.Model); ok && !m.IsQuitting() {
		selectedApp := m.GetSelectedApp()
		selectedFile := m.GetSelectedFile()

		if selectedApp != "" && selectedFile != "" {
			appConfig := registry.Apps[selectedApp]
			fileConfig := appConfig.Files[selectedFile]
			config.EditConfigFile(selectedApp, selectedFile, fileConfig)
		}
	}
}
