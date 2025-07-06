package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(0, 1).
			Width(80).
			Align(lipgloss.Center)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Padding(0, 1)

	focusedTabStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("51")).
			Padding(0, 1)

	activeFileStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")).
			Padding(0, 1)

	inactiveFileStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)

	missingFileStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Padding(0, 1)

	columnStyle = lipgloss.NewStyle().
			Width(0).
			Height(0)

	focusedColumnStyle = lipgloss.NewStyle().
				Width(0).
				Height(0).
				Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Padding(0, 1)
)

func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	m.updateDimensions()

	if m.expandedAppTab == -1 && len(m.appList) > 0 {
		m.expandedAppTab = 0
		m.activeAppTab = 0
		m.currentApp = m.appList[0]
		m.loadFileList()
	}

	// Show jump-to-line prompt if active (render like search bar at top, not footer)
	if m.jumpToLineMode {
		sections := []string{
			headerStyle.Render("🏗️HyDE User Config Manager"),
			"Goto line: " + m.jumpToLineInput + "█",
			m.renderMainContent(),
			m.renderDetailsBar(),
			m.renderFooter(),
		}
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	var sections []string

	header := headerStyle.Render("🏗️HyDE User Config Manager")
	sections = append(sections, header)

	mainContent := m.renderMainContent()
	sections = append(sections, mainContent)

	detailsBar := m.renderDetailsBar()
	sections = append(sections, detailsBar)

	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderDetailsBar() string {
	ColorBrightCyan := lipgloss.Color("51")
	ColorBrightBlack := lipgloss.Color("240")
	ColorBrightGreen := lipgloss.Color("82")
	ColorBrightRed := lipgloss.Color("196")

	barStyle := lipgloss.NewStyle().
		Foreground(ColorBrightCyan).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorBrightCyan).
		Padding(0, 1)
	sepStyle := lipgloss.NewStyle().Foreground(ColorBrightBlack)
	valueStyle := lipgloss.NewStyle().Foreground(ColorBrightBlack)
	okStyle := lipgloss.NewStyle().Foreground(ColorBrightGreen).Bold(true)
	errStyle := lipgloss.NewStyle().Foreground(ColorBrightRed).Bold(true)

	var info string

	activeAppTab := m.activeAppTab
	if m.focusArea == AppTabsFocus && (activeAppTab < 0 || activeAppTab >= len(m.appList)) && len(m.appList) > 0 {
		activeAppTab = 0
	}

	switch m.focusArea {
	case AppTabsFocus:
		if activeAppTab >= 0 && activeAppTab < len(m.appList) {
			appName := m.appList[activeAppTab]
			appConfig := m.registry.Apps[appName]
			if appConfig.Description != "" {
				info = valueStyle.Render(appConfig.Description)
			}
		}
	case FileTrayFocus:
		if m.activeFileTab >= 0 && m.activeFileTab < len(m.fileList) {
			fileName := m.fileList[m.activeFileTab]
			fileConfig := m.registry.Apps[m.currentApp].Files[fileName]
			if fileConfig.Description != "" {
				info = valueStyle.Render(fileConfig.Description)
			}
			if fileConfig.FileExists() {
				if info != "" {
					info += "  "
				}
				info += okStyle.Render("✓ Exists")
			} else {
				if info != "" {
					info += "  "
				}
				info += errStyle.Render("❌ Missing")
			}
		}
	case PreviewFocus:
		if m.activeFileTab >= 0 && m.activeFileTab < len(m.fileList) {
			fileName := m.fileList[m.activeFileTab]
			fileConfig := m.registry.Apps[m.currentApp].Files[fileName]
			if fileConfig.Description != "" {
				info = valueStyle.Render(fileConfig.Description)
			}
		}
	}

	if info == "" {
		info = sepStyle.Render("No selection. Use arrows to navigate.")
	}

	return barStyle.Width(m.windowWidth - 5).Render(info)
}

func (m *Model) renderMainContent() string {
	var columns []string

	// Render columns, adding a vertical border only to the active (focused) panel
	appCol := m.renderAppColumnNoBorder()
	if m.focusArea == AppTabsFocus {
		appCol = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).BorderRight(true).BorderTop(false).BorderBottom(false).BorderForeground(lipgloss.Color("51")).Bold(true).Width(m.tabWidth).Height(m.windowHeight-8).Render(appCol)
	}
	columns = append(columns, appCol)

	fileColumnPresent := false
	if m.expandedAppTab != -1 {
		fileCol := m.renderFileColumnNoBorder()
		if m.focusArea == FileTrayFocus {
			fileCol = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).BorderRight(true).BorderTop(false).BorderBottom(false).BorderForeground(lipgloss.Color("51")).Bold(true).Width(m.trayWidth).Height(m.windowHeight-8).Render(fileCol)
		}
		columns = append(columns, fileCol)
		fileColumnPresent = true
	}

	usedWidth := m.tabWidth
	if fileColumnPresent {
		usedWidth += m.trayWidth
	}

	previewWidth := m.windowWidth - usedWidth - 10
	if previewWidth < 10 {
		previewWidth = 10
	}

	parentHeight := m.windowHeight - 8
	previewCol := m.renderPreviewColumnWithWidthAndHeight(previewWidth, parentHeight)
	if m.focusArea == PreviewFocus {
		previewCol = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).BorderRight(true).BorderTop(false).BorderBottom(false).BorderForeground(lipgloss.Color("51")).Bold(true).Width(previewWidth).Height(parentHeight).Render(previewCol)
	}
	columns = append(columns, previewCol)

	mainBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Width(m.windowWidth - 2).
		Height(parentHeight)

	row := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	return mainBoxStyle.Render(row)
}

// Updated: renderPreviewColumnWithWidthAndHeight
func (m *Model) renderPreviewColumnWithWidthAndHeight(width, height int) string {
	icon := "🔎"
	header := fmt.Sprintf("%s Preview", icon)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	headerLine := headerStyle.Render(header)
	separatorLine := strings.Repeat("─", width-2)

	// Top elements: header and separator
	topElements := []string{headerLine, separatorLine}

	// Search bar or jump-to-line prompt (at top, if active)
	if m.searchMode && m.focusArea == PreviewFocus {
		searchBar := fmt.Sprintf("🔍 %s█", m.searchQuery)
		topElements = append(topElements, searchBar, "")
	}
	if m.jumpToLineMode {
		jumpBar := fmt.Sprintf("Goto line: %s█", m.jumpToLineInput)
		topElements = append(topElements, jumpBar, "")
	}

	// Always render all top elements, never collapse them
	topLines := len(topElements)
	maxLines := height
	contentLinesAvailable := maxLines - topLines
	if contentLinesAvailable < 0 {
		contentLinesAvailable = 0
	}

	// Prepare preview content block
	var contentBlock string
	if m.expandedAppTab != -1 && len(m.fileList) > 0 && m.activeFileTab < len(m.fileList) {
		fileName := m.fileList[m.activeFileTab]
		m.updatePreview(fileName)
		contentBlock = m.previewViewport.View()
	} else {
		contentBlock = ""
	}

	// Highlight regex matches in preview for both searchMode and n/N navigation
	var highlightQuery string
	if m.searchMode && m.focusArea == PreviewFocus && m.searchQuery != "" {
		highlightQuery = m.searchQuery
	} else if m.searchActive && m.focusArea == PreviewFocus && m.previewSearchBuffer != "" {
		highlightQuery = m.previewSearchBuffer
	}
	if highlightQuery != "" && contentBlock != "" {
		query := highlightQuery
		var re *regexp.Regexp
		var err error
		re, err = regexp.Compile("(?i)" + query)
		if err != nil {
			query = regexp.QuoteMeta(query)
			re = regexp.MustCompile("(?i)" + query)
		}
		indices := re.FindAllStringIndex(contentBlock, -1)
		current := m.previewMatchIndex
		var b strings.Builder
		last := 0
		for i, idx := range indices {
			b.WriteString(contentBlock[last:idx[0]])
			match := contentBlock[idx[0]:idx[1]]
			if i == current {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("226")).Bold(true).Render(match))
			} else {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Underline(true).Render(match))
			}
			last = idx[1]
		}
		b.WriteString(contentBlock[last:])
		contentBlock = b.String()
	}

	// Split content block into lines
	contentLines := []string{}
	if contentBlock != "" {
		contentLines = strings.Split(contentBlock, "\n")
	}

	// Only show as much content as fits below the top elements
	if len(contentLines) > contentLinesAvailable {
		contentLines = contentLines[:contentLinesAvailable]
	} else if len(contentLines) < contentLinesAvailable {
		for len(contentLines) < contentLinesAvailable {
			contentLines = append(contentLines, "")
		}
	}

	finalLines := append(topElements, contentLines...)
	for len(finalLines) < maxLines {
		finalLines = append(finalLines, "")
	}

	return lipgloss.NewStyle().Width(width).Height(maxLines).Render(strings.Join(finalLines, "\n"))
}

// Borderless versions of the column renderers
func (m *Model) renderAppColumnNoBorder() string {
	var content []string

	icon := "⚙️"
	header := fmt.Sprintf("%s Apps", icon)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	content = append(content, headerStyle.Render(header))
	content = append(content, strings.Repeat("─", m.tabWidth-2))

	if m.searchMode && m.focusArea == AppTabsFocus {
		searchBar := fmt.Sprintf("🔍 %s█", m.searchQuery)
		content = append(content, searchBar)
		content = append(content, "")
	}

	displayList := m.appList
	if m.searchMode && len(m.filteredApps) > 0 && m.focusArea == AppTabsFocus {
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

	// Always pad to parent height so parent box does not shrink/grow with content
	maxHeight := m.windowHeight - 8
	for len(content) < maxHeight {
		content = append(content, "")
	}

	columnContent := strings.Join(content, "\n")
	return lipgloss.NewStyle().Width(m.tabWidth).Height(maxHeight).Render(columnContent)
}

func (m *Model) renderFileColumnNoBorder() string {
	var content []string

	appConfig := m.registry.Apps[m.currentApp]
	icon := appConfig.Icon
	if icon == "" {
		icon = "⚙️"
	}
	header := fmt.Sprintf("%s Files", icon)
	content = append(content, header)
	content = append(content, strings.Repeat("─", m.trayWidth-2))

	if m.searchMode && m.focusArea == FileTrayFocus {
		searchBar := fmt.Sprintf("🔍 %s█", m.searchQuery)
		content = append(content, searchBar)
		content = append(content, "")
	}

	displayList := m.fileList
	if m.searchMode && len(m.filteredFiles) > 0 && m.focusArea == FileTrayFocus {
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

	// Always pad to parent height so parent box does not shrink/grow with content
	maxHeight := m.windowHeight - 8
	for len(content) < maxHeight {
		content = append(content, "")
	}

	columnContent := strings.Join(content, "\n")
	return lipgloss.NewStyle().Width(m.trayWidth).Height(maxHeight).Render(columnContent)
}

// Helper for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *Model) renderPreviewColumnWithWidthNoBorder(width int) string {
	icon := "🔎"
	header := fmt.Sprintf("%s Preview", icon)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	headerLine := headerStyle.Render(header)
	separatorLine := strings.Repeat("─", width-2)

	// Top elements: header and separator
	topElements := []string{headerLine, separatorLine}

	// Search bar or jump-to-line prompt (at top, if active)
	if m.searchMode && m.focusArea == PreviewFocus {
		searchBar := fmt.Sprintf("🔍 %s█", m.searchQuery)
		topElements = append(topElements, searchBar, "")
	}
	if m.jumpToLineMode {
		jumpBar := fmt.Sprintf("Goto line: %s█", m.jumpToLineInput)
		topElements = append(topElements, jumpBar, "")
	}

	// Calculate how many lines are reserved for top elements
	topLines := len(topElements)

	// Prepare preview content block
	var contentBlock string
	if m.expandedAppTab != -1 && len(m.fileList) > 0 && m.activeFileTab < len(m.fileList) {
		fileName := m.fileList[m.activeFileTab]
		m.updatePreview(fileName)
		contentBlock = m.previewViewport.View()
	} else {
		contentBlock = ""
	}

	// Highlight regex matches in preview for both searchMode and n/N navigation
	var highlightQuery string
	if m.searchMode && m.focusArea == PreviewFocus && m.searchQuery != "" {
		highlightQuery = m.searchQuery
	} else if m.searchActive && m.focusArea == PreviewFocus && m.previewSearchBuffer != "" {
		highlightQuery = m.previewSearchBuffer
	}
	if highlightQuery != "" && contentBlock != "" {
		query := highlightQuery
		var re *regexp.Regexp
		var err error
		re, err = regexp.Compile("(?i)" + query)
		if err != nil {
			query = regexp.QuoteMeta(query)
			re = regexp.MustCompile("(?i)" + query)
		}
		indices := re.FindAllStringIndex(contentBlock, -1)
		current := m.previewMatchIndex
		var b strings.Builder
		last := 0
		for i, idx := range indices {
			b.WriteString(contentBlock[last:idx[0]])
			match := contentBlock[idx[0]:idx[1]]
			if i == current {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("226")).Bold(true).Render(match))
			} else {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Underline(true).Render(match))
			}
			last = idx[1]
		}
		b.WriteString(contentBlock[last:])
		contentBlock = b.String()
	}

	// Split content block into lines
	contentLines := []string{}
	if contentBlock != "" {
		contentLines = strings.Split(contentBlock, "\n")
	}

	// Calculate available lines for content
	maxLines := m.windowHeight - 8
	contentLinesAvailable := maxLines - topLines
	if contentLinesAvailable < 0 {
		contentLinesAvailable = 0
	}

	// Truncate content lines if needed
	if len(contentLines) > contentLinesAvailable {
		contentLines = contentLines[:contentLinesAvailable]
	} else if len(contentLines) < contentLinesAvailable {
		for len(contentLines) < contentLinesAvailable {
			contentLines = append(contentLines, "")
		}
	}

	// Compose final lines: always show all top elements, then as much content as fits
	finalLines := append(topElements, contentLines...)
	// Pad to maxLines if needed
	for len(finalLines) < maxLines {
		finalLines = append(finalLines, "")
	}

	return lipgloss.NewStyle().Width(width).Height(maxLines).Render(strings.Join(finalLines, "\n"))
}

func (m *Model) renderPreviewColumn() string {
	return m.renderPreviewColumnWithWidthNoBorder(m.previewWidth)
}

func countWrappedLines(s string, width int) int {
	lines := strings.Split(s, "\n")
	count := 0
	for _, line := range lines {
		if width <= 0 {
			count++
			continue
		}
		visualLen := lipgloss.Width(line)
		if visualLen == 0 {
			count++
			continue
		}
		count += (visualLen-1)/width + 1
	}
	return count
}

func (m *Model) renderFooter() string {
	var statusItems []string

	if m.searchMode {
		// Only show search status in footer for AppTabsFocus and FileTrayFocus
		if m.focusArea == AppTabsFocus {
			statusItems = append(statusItems, fmt.Sprintf("Search apps: %s█", m.searchQuery))
			statusItems = append(statusItems, " Enter: confirm")
			statusItems = append(statusItems, " Esc: cancel")
		} else if m.focusArea == FileTrayFocus {
			statusItems = append(statusItems, fmt.Sprintf("Search files: %s█", m.searchQuery))
			statusItems = append(statusItems, " Enter: confirm")
			statusItems = append(statusItems, " Esc: cancel")
		}
	} else {
		switch m.focusArea {
		case AppTabsFocus:
			statusItems = append(statusItems, "↑/↓: navigate")
			statusItems = append(statusItems, " Enter/Space: expand")
		case FileTrayFocus:
			statusItems = append(statusItems, "↑/↓: navigate")
			statusItems = append(statusItems, " Enter: select")
			statusItems = append(statusItems, " ←: back to apps")
		case PreviewFocus:
			statusItems = append(statusItems, "PgUp/PgDn: scroll")
			statusItems = append(statusItems, " ←: back to files")
		}
		statusItems = append(statusItems, " Tab: cycle focus")
		statusItems = append(statusItems, " /: search")
		statusItems = append(statusItems, " q: quit")
	}

	statusText := strings.Join(statusItems, " ")
	return footerStyle.Width(m.windowWidth).Render(statusText)
}
