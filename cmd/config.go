package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"hydectl/internal/config"
	"hydectl/internal/tui"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Interactive configuration file editor",
	Long:  `Open an interactive selector to edit application configuration files with pre/post hooks.`,
	Run:   runConfigCommand,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfigCommand(cmd *cobra.Command, args []string) {
	registry, err := config.LoadConfigRegistry()
	if err != nil {
		fmt.Printf("Error loading config registry: %v\n", err)
		return
	}

	// If no apps in registry, show helpful message
	if len(registry.Apps) == 0 {
		fmt.Println("No applications found in config registry.")
		fmt.Println("Please add applications to your config-registry.toml file.")
		return
	}

	// Initialize the TUI model
	model := tui.NewModel(registry)

	// Start the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		return
	}

	// Handle the final result
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
