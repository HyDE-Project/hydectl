package tui

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"hydectl/internal/config"
)

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

func (m *Model) renderAppSelectionWithPreview() string {
	var s strings.Builder

	totalWidth := m.windowWidth
	if totalWidth < 80 {
		totalWidth = 80
	}
	listWidth := totalWidth / 2
	previewWidth := totalWidth - listWidth - 3

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

	if m.searchMode {
		s.WriteString(ColorBrightYellow + "❯ " + ColorReset + ColorBold + "Search: " + ColorReset)
		s.WriteString(ColorBrightGreen + m.searchQuery + "█" + ColorReset + "\n")
		s.WriteString(ColorBrightBlack + "  " + fmt.Sprintf("%d/%d matches", len(m.filteredList), len(m.appList)) + ColorReset + "\n\n")
	} else {
		s.WriteString(ColorBrightYellow + "❯ " + ColorReset + ColorBold + "Select an application to configure " + ColorReset)
		s.WriteString(ColorBrightBlack + "(press " + ColorReset + ColorBrightYellow + "/" + ColorReset + ColorBrightBlack + " to search)" + ColorReset + "\n\n")
	}

	var displayList []string
	if m.searchMode {
		displayList = m.filteredList
	} else {
		displayList = m.appList
	}

	var currentApp string
	if len(displayList) > 0 && m.cursor < len(displayList) {
		currentApp = displayList[m.cursor]
	}

	leftPaneLines := m.buildAppList(displayList, listWidth)

	var rightPaneLines []string
	if currentApp != "" {
		rightPaneLines = m.buildAppPreviewLines(currentApp, previewWidth)
	}

	maxLines := len(leftPaneLines)
	if len(rightPaneLines) > maxLines {
		maxLines = len(rightPaneLines)
	}

	for i := 0; i < maxLines; i++ {

		var leftLine string
		if i < len(leftPaneLines) {
			leftLine = leftPaneLines[i]
		} else {
			leftLine = strings.Repeat(" ", listWidth)
		}

		var rightLine string
		if i < len(rightPaneLines) {
			rightLine = rightPaneLines[i]
		} else {
			rightLine = strings.Repeat(" ", previewWidth)
		}

		s.WriteString(leftLine)
		s.WriteString(ColorBrightBlack + " │ " + ColorReset)
		s.WriteString(rightLine)
		s.WriteString("\n")
	}

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

func (m *Model) buildAppList(displayList []string, width int) []string {
	var lines []string
	maxDisplayItems := 15

	for i := 0; i < maxDisplayItems && i < len(displayList); i++ {
		app := displayList[i]
		appConfig := m.registry.Apps[app]
		icon := appConfig.Icon
		if icon == "" {
			icon = "⚙️"
		}

		var line string
		if i == m.cursor {

			desc := appConfig.Description
			maxDescLen := width - len(app) - 8
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

			desc := appConfig.Description
			maxDescLen := width - len(app) - 8
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

func (m *Model) buildAppPreviewLines(appName string, width int) []string {
	appConfig := m.registry.Apps[appName]
	var lines []string

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

	var files []string
	for fileName := range appConfig.Files {
		files = append(files, fileName)
	}
	sort.Strings(files)

	maxPreviewFiles := 10
	for i, fileName := range files {
		if i >= maxPreviewFiles {
			remainingCount := len(files) - maxPreviewFiles
			remainingLine := ColorBrightBlack + fmt.Sprintf("... and %d more files", remainingCount) + ColorReset
			remainingLine += strings.Repeat(" ", width-len(fmt.Sprintf("... and %d more files", remainingCount)))
			lines = append(lines, remainingLine)
			break
		}

		fileConfig := appConfig.Files[fileName]

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

		var maxFileLineLen int
		if exists {
			maxFileLineLen = len(fileName) + len(fileConfig.Description) + 5
		} else {
			maxFileLineLen = len(fileName) + len(fileConfig.Description) + 15
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

		displayLen := len(fileName) + len(fileConfig.Description) + 5
		if !exists {
			displayLen += 10
		}
		if displayLen < width {
			fileLine += strings.Repeat(" ", width-displayLen)
		}
		lines = append(lines, fileLine)

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

func (m *Model) buildFileList(displayList []string, width int) []string {
	var lines []string
	appConfig := m.registry.Apps[m.currentApp]

	for i, fileName := range displayList {
		fileConfig := appConfig.Files[fileName]
		fileExists := m.fileExists[fileName]

		var line string
		if i == m.cursor {

			if fileExists {
				desc := fileConfig.Description
				maxDescLen := width - len(fileName) - 10
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

				desc := fileConfig.Description
				maxDescLen := width - len(fileName) - 15
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

			if fileExists {
				desc := fileConfig.Description
				maxDescLen := width - len(fileName) - 8
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
				maxDescLen := width - len(fileName) - 13
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

func (m *Model) buildFilePreviewLines(fileName string, width int) []string {
	var lines []string
	appConfig := m.registry.Apps[m.currentApp]
	fileConfig := appConfig.Files[fileName]

	headerText := fmt.Sprintf("📄 %s", fileName)
	headerLine := ColorBrightCyan + headerText + ColorReset
	headerLine += strings.Repeat(" ", width-len(headerText))
	lines = append(lines, headerLine)

	separatorLine := ColorBrightBlack + strings.Repeat("─", width) + ColorReset
	lines = append(lines, separatorLine)

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

	lines = append(lines, strings.Repeat(" ", width))

	exists := fileConfig.FileExists()
	if exists {
		statusLine := ColorBrightGreen + "✓ File exists" + ColorReset
		statusLine += strings.Repeat(" ", width-len("✓ File exists"))
		lines = append(lines, statusLine)

		lines = append(lines, strings.Repeat(" ", width))
		previewHeaderLine := ColorBrightYellow + "File Content:" + ColorReset
		previewHeaderLine += strings.Repeat(" ", width-len("File Content:"))
		lines = append(lines, previewHeaderLine)

		separatorLine := ColorBrightBlack + strings.Repeat("─", width) + ColorReset
		lines = append(lines, separatorLine)

		expandedPath := config.ExpandPath(fileConfig.Path)
		allContent, _ := m.readFilePreviewWithScroll(expandedPath)

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

		lines = append(lines, strings.Repeat(" ", width))
		helpLine := ColorBrightBlack + "File will be created when selected" + ColorReset
		helpLine += strings.Repeat(" ", width-len("File will be created when selected"))
		lines = append(lines, helpLine)
	}

	return lines
}

func (m *Model) readFilePreview(filePath string, maxLines int) []string {
	var lines []string

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

	tempScanner := bufio.NewScanner(file)
	for tempScanner.Scan() {
		totalLines++
		if totalLines > maxLines*2 {
			break
		}
	}

	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()

		cleanLine := strings.Map(func(r rune) rune {
			if r < 32 && r != '\t' {
				return -1
			}
			return r
		}, line)
		lines = append(lines, cleanLine)
		lineCount++
	}

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

func (m *Model) readFilePreviewWithScroll(filePath string) ([]string, int) {
	var lines []string

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

	for scanner.Scan() {
		line := scanner.Text()

		cleanLine := strings.Map(func(r rune) rune {
			if r < 32 && r != '\t' {
				return -1
			}
			return r
		}, line)
		lines = append(lines, cleanLine)
		lineCount++

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

func (m *Model) renderFileSelection() string {
	var s strings.Builder

	totalWidth := m.windowWidth
	if totalWidth < 80 {
		totalWidth = 80
	}
	listWidth := totalWidth / 2
	previewWidth := totalWidth - listWidth - 3

	appConfig := m.registry.Apps[m.currentApp]
	icon := appConfig.Icon
	if icon == "" {
		icon = "⚙️"
	}

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

	if m.searchMode {
		s.WriteString(ColorBrightYellow + "❯ " + ColorReset + ColorBold + "Search files: " + ColorReset)
		s.WriteString(ColorBrightGreen + m.searchQuery + "█" + ColorReset + "\n")
		s.WriteString(ColorBrightBlack + "  " + fmt.Sprintf("%d/%d matches", len(m.filteredList), len(m.fileList)) + ColorReset + "\n\n")
	} else {
		s.WriteString(ColorBrightYellow + "❯ " + ColorReset + ColorBold + "Select a configuration file " + ColorReset)
		s.WriteString(ColorBrightBlack + "(press " + ColorReset + ColorBrightYellow + "/" + ColorReset + ColorBrightBlack + " to search)" + ColorReset + "\n\n")
	}

	var displayList []string
	if m.searchMode {
		displayList = m.filteredList
	} else {
		displayList = m.fileList
	}

	var currentFile string
	if len(displayList) > 0 && m.cursor < len(displayList) {
		currentFile = displayList[m.cursor]
	}

	leftPaneLines := m.buildFileList(displayList, listWidth)

	var rightPaneContent string
	if currentFile != "" {
		rightPaneLines := m.buildFilePreviewLines(currentFile, previewWidth)
		rightPaneContent = strings.Join(rightPaneLines, "\n")

		m.previewViewport.Width = previewWidth
		m.previewViewport.Height = 25
		m.previewViewport.SetContent(rightPaneContent)
	}

	maxLines := len(leftPaneLines)
	if maxLines < 25 {
		maxLines = 25
	}

	scrollbar := m.renderScrollbar(25, m.previewViewport.TotalLineCount(), m.previewViewport.YOffset)

	for i := 0; i < maxLines; i++ {

		var leftLine string
		if i < len(leftPaneLines) {
			leftLine = leftPaneLines[i]
		} else {
			leftLine = strings.Repeat(" ", listWidth)
		}

		var rightLine string
		if currentFile != "" && i < len(strings.Split(m.previewViewport.View(), "\n")) {
			viewportLines := strings.Split(m.previewViewport.View(), "\n")
			if i < len(viewportLines) {
				rightLine = viewportLines[i]

				if len(rightLine) < previewWidth {
					rightLine += strings.Repeat(" ", previewWidth-len(rightLine))
				}
			} else {
				rightLine = strings.Repeat(" ", previewWidth)
			}
		} else {
			rightLine = strings.Repeat(" ", previewWidth)
		}

		var scrollbarLine string
		if i < len(scrollbar) {
			scrollbarLine = scrollbar[i]
		} else {
			scrollbarLine = " "
		}

		s.WriteString(leftLine)
		s.WriteString(ColorBrightBlack + " │ " + ColorReset)
		s.WriteString(rightLine)
		s.WriteString(ColorBrightBlack + " " + ColorReset)
		s.WriteString(scrollbarLine)
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ColorBrightBlack + strings.Repeat("─", totalWidth) + ColorReset + "\n")

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

// renderScrollbar creates a visual scrollbar for the preview pane
func (m *Model) renderScrollbar(height int, totalLines int, currentOffset int) []string {
	var scrollbar []string

	if totalLines <= height {
		// No scrollbar needed if content fits
		for i := 0; i < height; i++ {
			scrollbar = append(scrollbar, " ")
		}
		return scrollbar
	}

	// Calculate scrollbar thumb position and size
	thumbSize := max(1, (height*height)/totalLines)
	if thumbSize > height {
		thumbSize = height
	}

	// Calculate thumb position
	scrollRatio := float64(currentOffset) / float64(totalLines-height)
	if scrollRatio < 0 {
		scrollRatio = 0
	}
	if scrollRatio > 1 {
		scrollRatio = 1
	}

	thumbPos := int(scrollRatio * float64(height-thumbSize))
	if thumbPos < 0 {
		thumbPos = 0
	}
	if thumbPos+thumbSize > height {
		thumbPos = height - thumbSize
	}

	// Build scrollbar
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			scrollbar = append(scrollbar, ColorBrightBlue+"█"+ColorReset) // Thumb
		} else if i == 0 || i == height-1 {
			scrollbar = append(scrollbar, ColorBrightBlack+"┃"+ColorReset) // Track ends
		} else {
			scrollbar = append(scrollbar, ColorBrightBlack+"│"+ColorReset) // Track
		}
	}

	return scrollbar
}

// renderSmoothScrollIndicator creates a smooth scroll position indicator
func (m *Model) renderSmoothScrollIndicator(currentLine, totalLines int) string {
	if totalLines <= 0 {
		return ""
	}

	percentage := (currentLine * 100) / totalLines
	if percentage > 100 {
		percentage = 100
	}

	// Create a visual progress bar
	barWidth := 20
	filled := (percentage * barWidth) / 100

	var bar strings.Builder
	bar.WriteString(ColorBrightCyan + "[" + ColorReset)

	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar.WriteString(ColorBrightGreen + "█" + ColorReset)
		} else if i == filled && percentage%5 != 0 {
			bar.WriteString(ColorBrightYellow + "▌" + ColorReset) // Half block for smooth transition
		} else {
			bar.WriteString(ColorBrightBlack + "░" + ColorReset)
		}
	}

	bar.WriteString(ColorBrightCyan + "]" + ColorReset)
	bar.WriteString(fmt.Sprintf(" %d%%", percentage))

	return bar.String()
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
