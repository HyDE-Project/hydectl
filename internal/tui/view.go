package tui

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"hydectl/internal/config"
)

// View renders the TUI with fzf-like preview
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	switch m.viewState {
	case AppSelectionView:
		s.WriteString(m.renderAppSelectionWithPreview())
	case FileSelectionView:
		s.WriteString(m.renderFileSelection())
	}

	return s.String()
}

// renderAppSelectionWithPreview renders the app selection view with preview pane
func (m *Model) renderAppSelectionWithPreview() string {
	var s strings.Builder

	// Calculate layout dimensions
	totalWidth := m.windowWidth
	if totalWidth < 80 {
		totalWidth = 80 // Minimum width
	}
	listWidth := totalWidth / 2
	previewWidth := totalWidth - listWidth - 3 // Account for separator

	// Header
	s.WriteString(ColorBrightCyan + "╭─" + strings.Repeat("─", totalWidth-4) + "─╮" + ColorReset + "\n")
	headerText := "🏗️  HyDE Configuration Manager"
	padding := (totalWidth - len(headerText) - 2) / 2
	if padding < 0 {
		padding = 0
	}
	s.WriteString(ColorBrightCyan + "│ " + ColorReset)
	s.WriteString(strings.Repeat(" ", padding))
	s.WriteString(ColorBrightWhite + ColorBold + headerText + ColorReset)
	s.WriteString(strings.Repeat(" ", totalWidth-len(headerText)-padding-4))
	s.WriteString(ColorBrightCyan + " │" + ColorReset + "\n")
	s.WriteString(ColorBrightCyan + "╰─" + strings.Repeat("─", totalWidth-4) + "─╯" + ColorReset + "\n")

	// Search bar
	if m.searchMode {
		s.WriteString(ColorBrightYellow + "❯ " + ColorReset + ColorBold + "Search: " + ColorReset)
		s.WriteString(ColorBrightGreen + m.searchQuery + "█" + ColorReset + "\n")
		s.WriteString(ColorBrightBlack + "  " + fmt.Sprintf("%d/%d matches", len(m.filteredList), len(m.appList)) + ColorReset + "\n\n")
	} else {
		s.WriteString(ColorBrightYellow + "❯ " + ColorReset + ColorBold + "Select an application to configure " + ColorReset)
		s.WriteString(ColorBrightBlack + "(press " + ColorReset + ColorBrightYellow + "/" + ColorReset + ColorBrightBlack + " to search)" + ColorReset + "\n\n")
	}

	// Prepare data for split panes
	var displayList []string
	if m.searchMode {
		displayList = m.filteredList
	} else {
		displayList = m.appList
	}

	// Get current app for preview
	var currentApp string
	if len(displayList) > 0 && m.cursor < len(displayList) {
		currentApp = displayList[m.cursor]
	}

	// Build left pane (application list)
	leftPaneLines := m.buildAppList(displayList, listWidth)

	// Build right pane (preview)
	var rightPaneLines []string
	if currentApp != "" {
		rightPaneLines = m.buildAppPreviewLines(currentApp, previewWidth)
	}

	// Render split panes side by side
	maxLines := len(leftPaneLines)
	if len(rightPaneLines) > maxLines {
		maxLines = len(rightPaneLines)
	}

	for i := 0; i < maxLines; i++ {
		// Left pane content
		var leftLine string
		if i < len(leftPaneLines) {
			leftLine = leftPaneLines[i]
		} else {
			leftLine = strings.Repeat(" ", listWidth)
		}

		// Right pane content
		var rightLine string
		if i < len(rightPaneLines) {
			rightLine = rightPaneLines[i]
		} else {
			rightLine = strings.Repeat(" ", previewWidth)
		}

		// Combine with separator
		s.WriteString(leftLine)
		s.WriteString(ColorBrightBlack + " │ " + ColorReset)
		s.WriteString(rightLine)
		s.WriteString("\n")
	}

	// Footer
	s.WriteString("\n")
	s.WriteString(ColorBrightBlack + strings.Repeat("─", totalWidth) + ColorReset + "\n")
	if m.searchMode {
		s.WriteString(ColorBrightBlue + "↑/↓" + ColorReset + ColorBrightBlack + " navigate  " + ColorReset)
		s.WriteString(ColorBrightGreen + "Enter" + ColorReset + ColorBrightBlack + " select  " + ColorReset)
		s.WriteString(ColorBrightYellow + "Esc" + ColorReset + ColorBrightBlack + " exit search  " + ColorReset)
		s.WriteString(ColorBrightRed + "Ctrl+C" + ColorReset + ColorBrightBlack + " quit" + ColorReset)
	} else {
		s.WriteString(ColorBrightBlue + "↑/↓" + ColorReset + ColorBrightBlack + " navigate  " + ColorReset)
		s.WriteString(ColorBrightGreen + "Enter" + ColorReset + ColorBrightBlack + " select  " + ColorReset)
		s.WriteString(ColorBrightYellow + "/" + ColorReset + ColorBrightBlack + " search  " + ColorReset)
		s.WriteString(ColorBrightRed + "q" + ColorReset + ColorBrightBlack + " quit" + ColorReset)
	}

	return s.String()
}

// buildAppList creates the left pane lines for the application list
func (m *Model) buildAppList(displayList []string, width int) []string {
	var lines []string
	maxDisplayItems := 15 // Limit number of items shown

	for i := 0; i < maxDisplayItems && i < len(displayList); i++ {
		app := displayList[i]
		appConfig := m.registry.Apps[app]
		icon := appConfig.Icon
		if icon == "" {
			icon = "⚙️"
		}

		var line string
		if i == m.cursor {
			// Selected item with highlight
			desc := appConfig.Description
			maxDescLen := width - len(app) - 8 // Account for icon, arrow, spacing
			if len(desc) > maxDescLen {
				desc = desc[:maxDescLen-3] + "..."
			}

			line = fmt.Sprintf("%s▶ %s%s %s %s- %s%s",
				ColorBrightMagenta,
				ColorBlack+BgBrightCyan,
				icon,
				app,
				ColorReset+ColorBrightBlack,
				desc,
				ColorReset)
			lines = append(lines, line+strings.Repeat(" ", width-len(app)-len(desc)-8))
		} else {
			// Regular item with inline description
			desc := appConfig.Description
			maxDescLen := width - len(app) - 8 // Account for icon, spacing
			if len(desc) > maxDescLen {
				desc = desc[:maxDescLen-3] + "..."
			}

			line = fmt.Sprintf("  %s%s %s%s %s- %s%s",
				ColorBrightGreen,
				icon,
				ColorBrightWhite,
				app,
				ColorBrightBlack,
				desc,
				ColorReset)
			lines = append(lines, line+strings.Repeat(" ", width-len(app)-len(desc)-8))
		}
	}

	return lines
}

// buildAppPreviewLines creates the right pane lines for the application preview
func (m *Model) buildAppPreviewLines(appName string, width int) []string {
	appConfig := m.registry.Apps[appName]
	var lines []string

	// Preview header
	icon := appConfig.Icon
	if icon == "" {
		icon = "⚙️"
	}

	headerText := fmt.Sprintf("%s %s Configuration Files", icon, appName)
	headerLine := ColorBrightCyan + headerText + ColorReset
	headerLine += strings.Repeat(" ", width-len(headerText))
	lines = append(lines, headerLine)

	separatorLine := ColorBrightBlack + strings.Repeat("─", width) + ColorReset
	lines = append(lines, separatorLine)

	// List configuration files
	var files []string
	for fileName := range appConfig.Files {
		files = append(files, fileName)
	}
	sort.Strings(files)

	maxPreviewFiles := 10 // Limit preview files
	for i, fileName := range files {
		if i >= maxPreviewFiles {
			remainingCount := len(files) - maxPreviewFiles
			remainingLine := ColorBrightBlack + fmt.Sprintf("... and %d more files", remainingCount) + ColorReset
			remainingLine += strings.Repeat(" ", width-len(fmt.Sprintf("... and %d more files", remainingCount)))
			lines = append(lines, remainingLine)
			break
		}

		fileConfig := appConfig.Files[fileName]

		// Check if file exists
		exists := fileConfig.FileExists()

		var fileLine string
		if exists {
			fileLine = fmt.Sprintf("%s📄 %s%s - %s%s",
				ColorBrightBlue,
				ColorBrightWhite,
				fileName,
				ColorBrightBlack,
				fileConfig.Description)
		} else {
			fileLine = fmt.Sprintf("%s❌ %s%s - %s (missing)%s",
				ColorDim,
				fileName,
				ColorDim,
				fileConfig.Description,
				ColorReset)
		}

		// Truncate if too long
		var maxFileLineLen int
		if exists {
			maxFileLineLen = len(fileName) + len(fileConfig.Description) + 5
		} else {
			maxFileLineLen = len(fileName) + len(fileConfig.Description) + 15 // for " (missing)"
		}

		if maxFileLineLen > width {
			maxDesc := width - len(fileName) - 15
			if maxDesc > 0 {
				if exists {
					fileLine = fmt.Sprintf("%s📄 %s%s - %s%s...%s",
						ColorBrightBlue,
						ColorBrightWhite,
						fileName,
						ColorBrightBlack,
						fileConfig.Description[:maxDesc],
						ColorReset)
				} else {
					fileLine = fmt.Sprintf("%s❌ %s%s - %s... (missing)%s",
						ColorDim,
						fileName,
						ColorDim,
						fileConfig.Description[:maxDesc],
						ColorReset)
				}
			}
		} else {
			fileLine += ColorReset
		}

		// Pad to full width
		displayLen := len(fileName) + len(fileConfig.Description) + 5
		if !exists {
			displayLen += 10 // " (missing)"
		}
		if displayLen < width {
			fileLine += strings.Repeat(" ", width-displayLen)
		}
		lines = append(lines, fileLine)

		// Show file path
		pathLine := fmt.Sprintf("    %s📁 %s%s", ColorBrightBlack, fileConfig.Path, ColorReset)
		pathDisplayLen := len(fileConfig.Path) + 7
		if pathDisplayLen > width {
			maxPath := width - 10
			if maxPath > 0 {
				pathLine = fmt.Sprintf("    %s📁 %s...%s", ColorBrightBlack, fileConfig.Path[:maxPath], ColorReset)
				pathDisplayLen = maxPath + 10
			}
		}

		if pathDisplayLen < width {
			pathLine += strings.Repeat(" ", width-pathDisplayLen)
		}
		lines = append(lines, pathLine)
	}

	return lines
}

// buildFileList creates the left pane lines for the file list
func (m *Model) buildFileList(displayList []string, width int) []string {
	var lines []string
	appConfig := m.registry.Apps[m.currentApp]

	for i, fileName := range displayList {
		fileConfig := appConfig.Files[fileName]
		fileExists := m.fileExists[fileName]

		var line string
		if i == m.cursor {
			// Selected file with highlight
			if fileExists {
				desc := fileConfig.Description
				maxDescLen := width - len(fileName) - 10 // Account for icon, arrow, spacing
				if len(desc) > maxDescLen {
					desc = desc[:maxDescLen-3] + "..."
				}

				line = fmt.Sprintf("%s▶ %s📄 %s %s- %s%s",
					ColorBrightMagenta,
					ColorBlack+BgBrightCyan,
					fileName,
					ColorReset+ColorBrightBlack,
					desc,
					ColorReset)
				lines = append(lines, line+strings.Repeat(" ", width-len(fileName)-len(desc)-10))
			} else {
				// Missing file - grayed out selection
				desc := fileConfig.Description
				maxDescLen := width - len(fileName) - 15 // Account for missing indicator
				if len(desc) > maxDescLen {
					desc = desc[:maxDescLen-3] + "..."
				}

				line = fmt.Sprintf("%s▶ %s❌ %s %s- %s (missing)%s",
					ColorDim,
					ColorDim+BgBrightBlack,
					fileName,
					ColorDim,
					desc,
					ColorReset)
				lines = append(lines, line+strings.Repeat(" ", width-len(fileName)-len(desc)-15))
			}
		} else {
			// Regular file item
			if fileExists {
				desc := fileConfig.Description
				maxDescLen := width - len(fileName) - 8 // Account for icon, spacing
				if len(desc) > maxDescLen {
					desc = desc[:maxDescLen-3] + "..."
				}

				line = fmt.Sprintf("  %s📄 %s%s %s- %s%s",
					ColorBrightBlue,
					ColorBrightWhite,
					fileName,
					ColorBrightBlack,
					desc,
					ColorReset)
				lines = append(lines, line+strings.Repeat(" ", width-len(fileName)-len(desc)-8))
			} else {
				desc := fileConfig.Description
				maxDescLen := width - len(fileName) - 13 // Account for missing indicator
				if len(desc) > maxDescLen {
					desc = desc[:maxDescLen-3] + "..."
				}

				line = fmt.Sprintf("  %s❌ %s%s %s- %s (missing)%s",
					ColorDim,
					fileName,
					ColorDim,
					ColorBrightBlack,
					desc,
					ColorReset)
				lines = append(lines, line+strings.Repeat(" ", width-len(fileName)-len(desc)-13))
			}
		}
	}

	return lines
}

// buildFilePreviewLines creates the right pane lines for file content preview
func (m *Model) buildFilePreviewLines(fileName string, width int) []string {
	var lines []string
	appConfig := m.registry.Apps[m.currentApp]
	fileConfig := appConfig.Files[fileName]

	// Preview header
	headerText := fmt.Sprintf("📄 %s", fileName)
	headerLine := ColorBrightCyan + headerText + ColorReset
	headerLine += strings.Repeat(" ", width-len(headerText))
	lines = append(lines, headerLine)

	separatorLine := ColorBrightBlack + strings.Repeat("─", width) + ColorReset
	lines = append(lines, separatorLine)

	// File info
	infoLine := fmt.Sprintf("%sDescription:%s %s", ColorBrightYellow, ColorReset, fileConfig.Description)
	if len(infoLine) > width {
		infoLine = infoLine[:width-3] + "..."
	}
	infoLine += strings.Repeat(" ", width-len(infoLine))
	lines = append(lines, infoLine)

	pathLine := fmt.Sprintf("%sPath:%s %s", ColorBrightYellow, ColorReset, fileConfig.Path)
	if len(pathLine) > width {
		pathLine = pathLine[:width-3] + "..."
	}
	pathLine += strings.Repeat(" ", width-len(pathLine))
	lines = append(lines, pathLine)

	// Add empty line
	lines = append(lines, strings.Repeat(" ", width))

	// Check if file exists and show preview
	exists := fileConfig.FileExists()
	if exists {
		statusLine := ColorBrightGreen + "✓ File exists" + ColorReset
		statusLine += strings.Repeat(" ", width-len("✓ File exists"))
		lines = append(lines, statusLine)

		// Try to read file content preview
		lines = append(lines, strings.Repeat(" ", width)) // Empty line
		previewHeaderLine := ColorBrightYellow + "File Content:" + ColorReset
		previewHeaderLine += strings.Repeat(" ", width-len("File Content:"))
		lines = append(lines, previewHeaderLine)

		separatorLine := ColorBrightBlack + strings.Repeat("─", width) + ColorReset
		lines = append(lines, separatorLine)

		// Read and display file content
		expandedPath := config.ExpandPath(fileConfig.Path)
		allContent, _ := m.readFilePreviewWithScroll(expandedPath)

		// Display content (viewport will handle scrolling)
		for _, contentLine := range allContent {
			if len(contentLine) > width {
				contentLine = contentLine[:width-3] + "..."
			}
			displayLine := ColorBrightWhite + contentLine + ColorReset
			displayLine += strings.Repeat(" ", width-len(contentLine))
			lines = append(lines, displayLine)
		}
	} else {
		statusLine := ColorBrightRed + "❌ File does not exist" + ColorReset
		statusLine += strings.Repeat(" ", width-len("❌ File does not exist"))
		lines = append(lines, statusLine)

		lines = append(lines, strings.Repeat(" ", width)) // Empty line
		helpLine := ColorBrightBlack + "File will be created when selected" + ColorReset
		helpLine += strings.Repeat(" ", width-len("File will be created when selected"))
		lines = append(lines, helpLine)
	}

	return lines
}

// readFilePreview reads the first N lines from a file for preview
func (m *Model) readFilePreview(filePath string, maxLines int) []string {
	var lines []string

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []string{ColorDim + "File does not exist" + ColorReset}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return []string{ColorBrightRed + "Error reading file: " + err.Error() + ColorReset}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	totalLines := 0

	// First pass: count total lines for large files
	tempScanner := bufio.NewScanner(file)
	for tempScanner.Scan() {
		totalLines++
		if totalLines > maxLines*2 { // Stop counting if it's way too large
			break
		}
	}

	// Reset to beginning
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		// Remove any control characters and limit line length
		cleanLine := strings.Map(func(r rune) rune {
			if r < 32 && r != '\t' {
				return -1 // Remove control characters except tab
			}
			return r
		}, line)
		lines = append(lines, cleanLine)
		lineCount++
	}

	// If there are more lines, add an indicator
	if totalLines > maxLines {
		moreLines := totalLines - maxLines
		if totalLines > maxLines*2 {
			lines = append(lines, ColorBrightBlack+"... (many more lines, file is large)"+ColorReset)
		} else {
			lines = append(lines, ColorBrightBlack+fmt.Sprintf("... (%d more lines)", moreLines)+ColorReset)
		}
	}

	if err := scanner.Err(); err != nil {
		lines = append(lines, ColorBrightRed+"Error reading file: "+err.Error()+ColorReset)
	}

	if lineCount == 0 && totalLines == 0 {
		lines = append(lines, ColorBrightBlack+"(empty file)"+ColorReset)
	}

	return lines
}

// readFilePreviewWithScroll reads file content with scroll support and returns all content plus total line count
func (m *Model) readFilePreviewWithScroll(filePath string) ([]string, int) {
	var lines []string

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []string{ColorDim + "File does not exist" + ColorReset}, 1
	}

	file, err := os.Open(filePath)
	if err != nil {
		return []string{ColorBrightRed + "Error reading file: " + err.Error() + ColorReset}, 1
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	// Read all lines for scrolling
	for scanner.Scan() {
		line := scanner.Text()
		// Remove any control characters and limit line length
		cleanLine := strings.Map(func(r rune) rune {
			if r < 32 && r != '\t' {
				return -1 // Remove control characters except tab
			}
			return r
		}, line)
		lines = append(lines, cleanLine)
		lineCount++

		// Safety limit for very large files
		if lineCount > 10000 {
			lines = append(lines, ColorBrightBlack+"... (file too large, showing first 10000 lines)"+ColorReset)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		lines = append(lines, ColorBrightRed+"Error reading file: "+err.Error()+ColorReset)
	}

	if lineCount == 0 {
		lines = append(lines, ColorBrightBlack+"(empty file)"+ColorReset)
		return lines, 1
	}

	return lines, lineCount
}

// renderFileSelection renders the file selection view with preview pane
func (m *Model) renderFileSelection() string {
	var s strings.Builder

	// Calculate layout dimensions
	totalWidth := m.windowWidth
	if totalWidth < 80 {
		totalWidth = 80 // Minimum width
	}
	listWidth := totalWidth / 2
	previewWidth := totalWidth - listWidth - 3 // Account for separator

	appConfig := m.registry.Apps[m.currentApp]
	icon := appConfig.Icon
	if icon == "" {
		icon = "⚙️"
	}

	// Header
	s.WriteString(ColorBrightCyan + "╭─" + strings.Repeat("─", totalWidth-4) + "─╮" + ColorReset + "\n")
	headerText := fmt.Sprintf("%s %s Configuration", icon, m.currentApp)
	padding := (totalWidth - len(headerText) - 2) / 2
	if padding < 0 {
		padding = 0
	}
	s.WriteString(ColorBrightCyan + "│ " + ColorReset)
	s.WriteString(strings.Repeat(" ", padding))
	s.WriteString(ColorBrightWhite + ColorBold + headerText + ColorReset)
	s.WriteString(strings.Repeat(" ", totalWidth-len(headerText)-padding-4))
	s.WriteString(ColorBrightCyan + " │" + ColorReset + "\n")
	s.WriteString(ColorBrightCyan + "╰─" + strings.Repeat("─", totalWidth-4) + "─╯" + ColorReset + "\n")

	// Search bar for files
	if m.searchMode {
		s.WriteString(ColorBrightYellow + "❯ " + ColorReset + ColorBold + "Search files: " + ColorReset)
		s.WriteString(ColorBrightGreen + m.searchQuery + "█" + ColorReset + "\n")
		s.WriteString(ColorBrightBlack + "  " + fmt.Sprintf("%d/%d matches", len(m.filteredList), len(m.fileList)) + ColorReset + "\n\n")
	} else {
		s.WriteString(ColorBrightYellow + "❯ " + ColorReset + ColorBold + "Select a configuration file " + ColorReset)
		s.WriteString(ColorBrightBlack + "(press " + ColorReset + ColorBrightYellow + "/" + ColorReset + ColorBrightBlack + " to search)" + ColorReset + "\n\n")
	}

	// Prepare data for split panes
	var displayList []string
	if m.searchMode {
		displayList = m.filteredList
	} else {
		displayList = m.fileList
	}

	// Get current file for preview
	var currentFile string
	if len(displayList) > 0 && m.cursor < len(displayList) {
		currentFile = displayList[m.cursor]
	}

	// Build left pane (file list)
	leftPaneLines := m.buildFileList(displayList, listWidth)

	// Build right pane (file content preview with scrolling)
	var rightPaneContent string
	if currentFile != "" {
		rightPaneLines := m.buildFilePreviewLines(currentFile, previewWidth)
		rightPaneContent = strings.Join(rightPaneLines, "\n")

		// Update viewport with the file content
		m.previewViewport.Width = previewWidth
		m.previewViewport.Height = 25 // Available preview height
		m.previewViewport.SetContent(rightPaneContent)
	}

	// Render split panes side by side
	maxLines := len(leftPaneLines)
	if maxLines < 25 { // Minimum height for viewport
		maxLines = 25
	}

	for i := 0; i < maxLines; i++ {
		// Left pane content
		var leftLine string
		if i < len(leftPaneLines) {
			leftLine = leftPaneLines[i]
		} else {
			leftLine = strings.Repeat(" ", listWidth)
		}

		// Right pane content from viewport
		var rightLine string
		if currentFile != "" && i < len(strings.Split(m.previewViewport.View(), "\n")) {
			viewportLines := strings.Split(m.previewViewport.View(), "\n")
			if i < len(viewportLines) {
				rightLine = viewportLines[i]
				// Ensure proper width
				if len(rightLine) < previewWidth {
					rightLine += strings.Repeat(" ", previewWidth-len(rightLine))
				}
			} else {
				rightLine = strings.Repeat(" ", previewWidth)
			}
		} else {
			rightLine = strings.Repeat(" ", previewWidth)
		}

		// Combine with separator
		s.WriteString(leftLine)
		s.WriteString(ColorBrightBlack + " │ " + ColorReset)
		s.WriteString(rightLine)
		s.WriteString("\n")
	}

	// Footer with scroll indicators
	s.WriteString("\n")
	s.WriteString(ColorBrightBlack + strings.Repeat("─", totalWidth) + ColorReset + "\n")

	// Show scroll information if preview is scrollable
	scrollInfo := ""
	if m.previewViewport.TotalLineCount() > 25 && currentFile != "" {
		scrollInfo = fmt.Sprintf(" [%d/%d lines] ", m.previewViewport.YOffset+1, m.previewViewport.TotalLineCount())
	}

	if m.searchMode {
		s.WriteString(ColorBrightBlue + "↑/↓" + ColorReset + ColorBrightBlack + " navigate  " + ColorReset)
		s.WriteString(ColorBrightGreen + "Enter" + ColorReset + ColorBrightBlack + " select  " + ColorReset)
		s.WriteString(ColorBrightYellow + "Esc" + ColorReset + ColorBrightBlack + " exit search  " + ColorReset)
		s.WriteString(ColorBrightRed + "Ctrl+C" + ColorReset + ColorBrightBlack + " quit" + ColorReset)
	} else {
		s.WriteString(ColorBrightBlue + "↑/↓" + ColorReset + ColorBrightBlack + " navigate  " + ColorReset)
		s.WriteString(ColorBrightGreen + "Enter" + ColorReset + ColorBrightBlack + " select  " + ColorReset)
		s.WriteString(ColorBrightYellow + "/" + ColorReset + ColorBrightBlack + " search  " + ColorReset)
		s.WriteString(ColorBrightMagenta + "PgUp/PgDn/🖱️" + ColorReset + ColorBrightBlack + " scroll preview  " + ColorReset)
		s.WriteString(ColorBrightYellow + "←/h" + ColorReset + ColorBrightBlack + " back  " + ColorReset)
		s.WriteString(ColorBrightRed + "q" + ColorReset + ColorBrightBlack + " quit" + ColorReset)
	}

	if scrollInfo != "" {
		s.WriteString(ColorBrightCyan + scrollInfo + ColorReset)
	}

	return s.String()
}
