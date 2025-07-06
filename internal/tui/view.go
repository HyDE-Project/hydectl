package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Width(80).
			Align(lipgloss.Center)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("86")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	focusedTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Padding(0, 1)

	activeFileStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("226")).
			Padding(0, 1)

	inactiveFileStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	missingFileStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)
)

func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	m.updateDimensions()

	var sections []string

	header := headerStyle.Render("🏗️HyDE User Config Manager")
	sections = append(sections, header)

	mainContent := m.renderMainContent()
	sections = append(sections, mainContent)

	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderMainContent() string {
	var columns []string

	appColumn := m.renderAppColumn()
	columns = append(columns, appColumn)

	if m.expandedAppTab != -1 {
		fileColumn := m.renderFileColumn()
		columns = append(columns, fileColumn)
	}

	previewColumn := m.renderPreviewColumn()
	columns = append(columns, previewColumn)

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func (m *Model) renderAppColumn() string {
	var content []string

	if m.searchMode {
		searchBar := fmt.Sprintf("🔍 %s█", m.searchQuery)
		content = append(content, searchBar)
		content = append(content, "")
	}

	displayList := m.appList
	if m.searchMode && len(m.filteredApps) > 0 {
		displayList = m.filteredApps
	}

	for i, appName := range displayList {
		appConfig := m.registry.Apps[appName]
		icon := appConfig.Icon
		if icon == "" {
			icon = "⚙️"
		}

		var displayText string
		if m.expandedAppTab == i {
			displayText = fmt.Sprintf("▼ %s %s", icon, appName)
		} else {
			displayText = fmt.Sprintf("▶ %s %s", icon, appName)
		}

		var styled string
		if i == m.activeAppTab && m.focusArea == AppTabsFocus {
			styled = focusedTabStyle.Render(displayText)
		} else if i == m.activeAppTab {
			styled = activeTabStyle.Render(displayText)
		} else {
			styled = inactiveTabStyle.Render(displayText)
		}

		content = append(content, styled)
	}

	for len(content) < m.windowHeight-8 {
		content = append(content, "")
	}

	columnContent := strings.Join(content, "\n")
	return columnStyle.
		Width(m.tabWidth).
		Height(m.windowHeight - 8).
		Render(columnContent)
}

func (m *Model) renderFileColumn() string {
	if m.expandedAppTab == -1 || m.currentApp == "" {
		return ""
	}

	var content []string

	appConfig := m.registry.Apps[m.currentApp]
	icon := appConfig.Icon
	if icon == "" {
		icon = "⚙️"
	}
	header := fmt.Sprintf("%s Files", icon)
	content = append(content, header)
	content = append(content, strings.Repeat("─", m.trayWidth-2))

	displayList := m.fileList
	if m.searchMode && len(m.filteredFiles) > 0 {
		displayList = m.filteredFiles
	}

	for i, fileName := range displayList {
		exists := m.fileExists[fileName]

		var displayText string
		if exists {
			displayText = fmt.Sprintf("📄 %s", fileName)
		} else {
			displayText = fmt.Sprintf("❌ %s", fileName)
		}

		var styled string
		if i == m.activeFileTab && m.focusArea == FileTrayFocus {
			if exists {
				styled = activeFileStyle.Render(displayText)
			} else {
				styled = missingFileStyle.Render(displayText)
			}
		} else if i == m.activeFileTab {
			if exists {
				styled = activeFileStyle.Render(displayText)
			} else {
				styled = missingFileStyle.Render(displayText)
			}
		} else {
			if exists {
				styled = inactiveFileStyle.Render(displayText)
			} else {
				styled = missingFileStyle.Render(displayText)
			}
		}

		content = append(content, styled)
	}

	for len(content) < m.windowHeight-8 {
		content = append(content, "")
	}

	columnContent := strings.Join(content, "\n")
	return columnStyle.
		Width(m.trayWidth).
		Height(m.windowHeight - 8).
		Render(columnContent)
}

func (m *Model) renderPreviewColumn() string {
	var content string

	if m.expandedAppTab != -1 && len(m.fileList) > 0 && m.activeFileTab < len(m.fileList) {

		fileName := m.fileList[m.activeFileTab]
		m.updatePreview(fileName)
		content = m.previewViewport.View()
	} else {

		welcomeLines := []string{
			"",
			"Welcome to HyDE Config Manager",
			"",
			"← Select an app from the left panel",
			"Press Enter or Space to expand",
			"",
			"Navigation:",
			"↑/↓ or k/j - Move up/down",
			"←/→ or h/l - Move between panels",
			"Tab/Shift+Tab - Cycle focus",
			"Space/Enter - Expand/select",
			"/ - Search",
			"q - Quit",
		}
		content = strings.Join(welcomeLines, "\n")
	}

	return columnStyle.
		Width(m.previewWidth).
		Height(m.windowHeight - 8).
		Render(content)
}

func (m *Model) renderFooter() string {
	var statusItems []string

	if m.searchMode {
		statusItems = append(statusItems, fmt.Sprintf("Search: %s█", m.searchQuery))
		statusItems = append(statusItems, "Enter: confirm")
		statusItems = append(statusItems, "Esc: cancel")
	} else {

		switch m.focusArea {
		case AppTabsFocus:
			statusItems = append(statusItems, "↑/↓: navigate")
			statusItems = append(statusItems, "Enter/Space: expand")
		case FileTrayFocus:
			statusItems = append(statusItems, "↑/↓: navigate")
			statusItems = append(statusItems, "Enter: select")
			statusItems = append(statusItems, "←: back to apps")
		case PreviewFocus:
			statusItems = append(statusItems, "PgUp/PgDn: scroll")
			statusItems = append(statusItems, "←: back to files")
		}

		statusItems = append(statusItems, "Tab: cycle focus")
		statusItems = append(statusItems, "/: search")
		statusItems = append(statusItems, "q: quit")
	}

	statusText := strings.Join(statusItems, "")
	return footerStyle.Width(m.windowWidth).Render(statusText)
}
