package tui

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"fmt"

	chroma "github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hydectl/internal/config"
)

type FocusArea int

const (
	AppTabsFocus FocusArea = iota
	FileTrayFocus
	PreviewFocus
)

type Model struct {
	registry   *config.ConfigRegistry
	appList    []string
	fileList   []string
	fileExists map[string]bool

	activeAppTab   int
	expandedAppTab int
	activeFileTab  int
	focusArea      FocusArea

	windowWidth  int
	windowHeight int
	tabWidth     int
	trayWidth    int
	previewWidth int

	searchQuery   string
	searchMode    bool
	filteredApps  []string
	filteredFiles []string

	quitting     bool
	selectedFile string
	currentApp   string

	previewViewport  viewport.Model
	fileTrayViewport viewport.Model

	lastScrollTime time.Time

	highlightStyle      string
	previewMatchIndices []int // byte offsets of regex matches in preview
	previewMatchIndex   int   // current match index
}

func NewModel(registry *config.ConfigRegistry, highlightStyle string) *Model {
	var apps []string
	for appName := range registry.Apps {
		apps = append(apps, appName)
	}
	sort.Strings(apps)

	previewVp := viewport.New(60, 25)
	previewVp.YPosition = 0

	trayVp := viewport.New(30, 20)
	trayVp.YPosition = 0

	return &Model{
		registry:         registry,
		appList:          apps,
		fileExists:       make(map[string]bool),
		activeAppTab:     0,
		expandedAppTab:   -1,
		activeFileTab:    0,
		focusArea:        AppTabsFocus,
		windowWidth:      120,
		windowHeight:     30,
		tabWidth:         25,
		trayWidth:        35,
		previewWidth:     60,
		previewViewport:  previewVp,
		fileTrayViewport: trayVp,
		highlightStyle:   highlightStyle,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.updateDimensions()

	case tea.MouseMsg:
		if msg.Type == tea.MouseWheelUp || msg.Type == tea.MouseWheelDown {
			now := time.Now()
			if now.Sub(m.lastScrollTime) < 50*time.Millisecond {
				return m, nil
			}
			m.lastScrollTime = now
		}

		if msg.X < m.tabWidth {
			m.focusArea = AppTabsFocus
		} else if m.expandedAppTab != -1 && msg.X < m.tabWidth+m.trayWidth {
			m.focusArea = FileTrayFocus
		} else {
			m.focusArea = PreviewFocus
		}

		switch msg.Type {
		case tea.MouseWheelUp:
			if m.focusArea == AppTabsFocus {
				if m.activeAppTab > 0 {
					m.activeAppTab--
					m.expandAppTab(m.activeAppTab)
				}
			} else if m.focusArea == FileTrayFocus {
				if m.activeFileTab > 0 {
					m.activeFileTab--
					if len(m.fileList) > 0 {
						m.updatePreview(m.fileList[m.activeFileTab])
					}
				}
			} else if m.focusArea == PreviewFocus {
				m.previewViewport.ScrollUp(1)
			}
		case tea.MouseWheelDown:
			if m.focusArea == AppTabsFocus {
				if m.activeAppTab < len(m.appList)-1 {
					m.activeAppTab++
					m.expandAppTab(m.activeAppTab)
				}
			} else if m.focusArea == FileTrayFocus {
				if m.activeFileTab < len(m.fileList)-1 {
					m.activeFileTab++
					if len(m.fileList) > 0 {
						m.updatePreview(m.fileList[m.activeFileTab])
					}
				}
			} else if m.focusArea == PreviewFocus {
				m.previewViewport.ScrollDown(1)
			}
		case tea.MouseLeft:
			if msg.X < m.tabWidth {
				m.focusArea = AppTabsFocus
			} else if m.expandedAppTab != -1 && msg.X < m.tabWidth+m.trayWidth {
				m.focusArea = FileTrayFocus
			} else {
				m.focusArea = PreviewFocus
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.searchMode {
			if m.focusArea == PreviewFocus && (msg.String() == "n" || msg.String() == "N") {
				if len(m.previewMatchIndices) == 0 {
					return m, nil
				}
				if msg.String() == "n" {
					m.previewMatchIndex = (m.previewMatchIndex + 1) % len(m.previewMatchIndices)
				} else {
					m.previewMatchIndex = (m.previewMatchIndex - 1 + len(m.previewMatchIndices)) % len(m.previewMatchIndices)
				}
				// Scroll preview to match
				m.scrollPreviewToMatch()
				return m, nil
			}
			return m.handleSearchMode(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "/":
			m.searchMode = true
			m.searchQuery = ""
			m.updateFilteredLists()

		case "tab":
			m.cycleFocus(1)

		case "shift+tab":
			m.cycleFocus(-1)

		case "left", "h":

			if m.focusArea == FileTrayFocus {
				m.focusArea = AppTabsFocus
			} else if m.focusArea == PreviewFocus {
				m.focusArea = FileTrayFocus
			}

		case "right", "l":

			if m.focusArea == AppTabsFocus && m.expandedAppTab != -1 {
				m.focusArea = FileTrayFocus
			} else if m.focusArea == FileTrayFocus {
				m.focusArea = PreviewFocus
			}

		case "enter":
			return m.handleEnter()

		case "up", "k":
			if m.focusArea == AppTabsFocus {
				if m.activeAppTab > 0 {
					m.activeAppTab--
					m.expandAppTab(m.activeAppTab)
				}
			} else if m.focusArea == FileTrayFocus {
				if m.activeFileTab > 0 {
					m.activeFileTab--
					if len(m.fileList) > 0 {
						m.updatePreview(m.fileList[m.activeFileTab])
					}
				}
			} else if m.focusArea == PreviewFocus {
				m.previewViewport.ScrollUp(1)
			}

		case "down", "j":
			if m.focusArea == AppTabsFocus {
				if m.activeAppTab < len(m.appList)-1 {
					m.activeAppTab++
					m.expandAppTab(m.activeAppTab)
				}
			} else if m.focusArea == FileTrayFocus {
				if m.activeFileTab < len(m.fileList)-1 {
					m.activeFileTab++
					if len(m.fileList) > 0 {
						m.updatePreview(m.fileList[m.activeFileTab])
					}
				}
			} else if m.focusArea == PreviewFocus {
				m.previewViewport.ScrollDown(1)
			}
		}
	}

	return m, nil
}

func (m *Model) updateDimensions() {

	m.tabWidth = 25
	m.trayWidth = 35

	usedWidth := m.tabWidth + 2
	if m.expandedAppTab != -1 {
		usedWidth += m.trayWidth + 1
	}

	m.previewWidth = m.windowWidth - usedWidth
	if m.previewWidth < 30 {
		m.previewWidth = 30
	}

	contentHeight := m.windowHeight - 8
	if contentHeight < 10 {
		contentHeight = 10
	}

	m.previewViewport.Width = m.previewWidth
	m.previewViewport.Height = contentHeight
	m.fileTrayViewport.Width = m.trayWidth
	m.fileTrayViewport.Height = contentHeight
}

func (m *Model) cycleFocus(direction int) {
	areas := []FocusArea{AppTabsFocus}

	if m.expandedAppTab != -1 {
		areas = append(areas, FileTrayFocus)
		areas = append(areas, PreviewFocus)
	}

	currentIndex := 0
	for i, area := range areas {
		if area == m.focusArea {
			currentIndex = i
			break
		}
	}

	if direction > 0 {
		currentIndex = (currentIndex + 1) % len(areas)
	} else {
		currentIndex = (currentIndex - 1 + len(areas)) % len(areas)
	}

	m.focusArea = areas[currentIndex]
}

func (m *Model) expandAppTab(appIndex int) {
	if appIndex >= 0 && appIndex < len(m.appList) {
		m.expandedAppTab = appIndex
		m.currentApp = m.appList[appIndex]
		m.loadFileList()
		m.activeFileTab = 0

	}
}

func (m *Model) loadFileList() {
	if m.currentApp == "" {
		return
	}

	appConfig := m.registry.Apps[m.currentApp]
	var files []string
	for fileName := range appConfig.Files {
		files = append(files, fileName)
	}
	sort.Strings(files)

	m.fileList = files
	m.checkFileExists()

	if len(files) > 0 && m.activeFileTab < len(files) {
		m.updatePreview(files[m.activeFileTab])
	}
}

func (m *Model) checkFileExists() {
	if m.fileExists == nil {
		m.fileExists = make(map[string]bool)
	}

	if m.currentApp == "" {
		return
	}

	appConfig := m.registry.Apps[m.currentApp]
	for fileName, fileConfig := range appConfig.Files {
		m.fileExists[fileName] = fileConfig.FileExists()
	}
}

func highlightFileContent(displayName, realPath, content, styleName string) string {

	lexer := lexers.Match(realPath)

	if lexer == nil {
		name := strings.ToLower(realPath)
		if strings.Contains(name, "css") {
			lexer = lexers.Get("css")
		} else if strings.Contains(name, "toml") {
			lexer = lexers.Get("toml")
		} else if strings.Contains(name, "conf") || strings.Contains(name, "rc") {
			lexer = lexers.Get("ini")
		} else if strings.Contains(name, "json") {
			lexer = lexers.Get("json")
		} else if strings.Contains(name, "sh") || strings.Contains(name, "bash") || strings.Contains(name, "zsh") {
			lexer = lexers.Get("bash")
		} else if strings.Contains(name, "yaml") || strings.Contains(name, "yml") {
			lexer = lexers.Get("yaml")
		} else if strings.Contains(name, "lua") {
			lexer = lexers.Get("lua")
		} else if strings.Contains(name, "py") {
			lexer = lexers.Get("python")
		} else if strings.Contains(name, "js") {
			lexer = lexers.Get("javascript")
		} else if strings.Contains(name, "hypr") {
			lexer = lexers.Get("ini")
		}
	}

	if lexer == nil {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		logTuiDebug(fmt.Sprintf("[highlightFileContent] No lexer found for %s (realPath: %s)", displayName, realPath))
		return content
	}

	tryStyles := []string{styleName, "native", "github", "monokai", "dracula"}
	var style *chroma.Style
	var styleUsed string
	for _, s := range tryStyles {
		if s == "" {
			continue
		}
		style = styles.Get(s)
		if style != nil {
			styleUsed = s
			break
		}
	}
	if style == nil {
		style = styles.Fallback
		styleUsed = "fallback"
	}

	logTuiDebug(fmt.Sprintf("[highlightFileContent] File: %s | RealPath: %s | Lexer: %s | Style: %s", displayName, realPath, lexer.Config().Name, styleUsed))

	var buf bytes.Buffer
	err := quick.Highlight(&buf, content, lexer.Config().Name, "terminal256", style.Name)
	if err != nil {
		logTuiDebug(fmt.Sprintf("[highlightFileContent] Chroma error: %v", err))
		return content
	}
	return buf.String()
}

func logTuiDebug(msg string) {
	f, err := os.OpenFile("/tmp/hydectl-tui-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	f.WriteString("[" + timestamp + "] " + msg + "\n")
}

func (m *Model) updatePreview(fileName string) {
	if m.currentApp == "" || fileName == "" {
		return
	}

	appConfig := m.registry.Apps[m.currentApp]
	fileConfig, exists := appConfig.Files[fileName]
	if !exists {
		return
	}

	var contentLines []string
	if fileConfig.FileExists() {
		expandedPath := config.ExpandPath(fileConfig.Path)
		content, _ := m.readFileContent(expandedPath)
		joined := strings.Join(content, "\n")
		highlighted := highlightFileContent(fileName, expandedPath, joined, m.highlightStyle)
		contentLines = strings.Split(highlighted, "\n")
	} else {
		contentLines = []string{}
	}

	m.previewViewport.SetContent(strings.Join(contentLines, "\n"))
	m.previewMatchIndices = nil
	m.previewMatchIndex = 0
	if m.searchMode && m.focusArea == PreviewFocus && m.searchQuery != "" {
		m.updatePreviewMatches()
	}
}

func (m *Model) updatePreviewMatches() {
	content := m.previewViewport.View()
	m.previewMatchIndices = nil
	m.previewMatchIndex = 0
	if m.searchQuery == "" {
		return
	}
	indices := regexAllIndices(content, m.searchQuery)
	m.previewMatchIndices = indices
}

// regexAllIndices returns the start indices of all regex matches (case-insensitive, fallback to literal if invalid)
func regexAllIndices(text, pattern string) []int {
	var indices []int
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		// fallback to literal
		pattern = regexp.QuoteMeta(pattern)
		re = regexp.MustCompile("(?i)" + pattern)
	}
	matches := re.FindAllStringIndex(text, -1)
	for _, m := range matches {
		indices = append(indices, m[0])
	}
	return indices
}

func (m *Model) readFileContent(filePath string) ([]string, error) {

	content, _ := m.readFilePreviewWithScroll(filePath)
	return content, nil
}

func (m *Model) readFilePreviewWithScroll(filePath string) ([]string, int) {
	var lines []string

	ColorBrightBlack := lipgloss.Color("240")
	ColorBrightRed := lipgloss.Color("196")
	ColorDim := lipgloss.Color("245")

	sepStyle := lipgloss.NewStyle().Foreground(ColorBrightBlack)
	errStyle := lipgloss.NewStyle().Foreground(ColorBrightRed).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(ColorDim)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []string{dimStyle.Render("File does not exist")}, 1
	}

	file, err := os.Open(filePath)
	if err != nil {
		return []string{errStyle.Render("Error reading file: " + err.Error())}, 1
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
			lines = append(lines, sepStyle.Render("... (file too large, showing first 10000 lines)"))
			break
		}
	}

	if err := scanner.Err(); err != nil {
		lines = append(lines, errStyle.Render("Error reading file: "+err.Error()))
	}

	if lineCount == 0 {
		lines = append(lines, sepStyle.Render("(empty file)"))
		return lines, 1
	}

	return lines, lineCount
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.focusArea {
	case AppTabsFocus:
		if m.expandedAppTab != m.activeAppTab {
			m.expandAppTab(m.activeAppTab)
		}
		m.focusArea = FileTrayFocus
	case FileTrayFocus:
		if len(m.fileList) > 0 && m.activeFileTab < len(m.fileList) {
			fileName := m.fileList[m.activeFileTab]
			if m.canSelectFile(fileName) {
				m.selectedFile = fileName
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *Model) canSelectFile(fileName string) bool {
	exists, found := m.fileExists[fileName]
	return found && exists
}

func (m *Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchMode = false

		if m.focusArea == AppTabsFocus && len(m.filteredApps) > 0 {
			for i, app := range m.appList {
				if app == m.filteredApps[0] {
					m.activeAppTab = i
					m.expandAppTab(i)
					m.focusArea = FileTrayFocus // Move focus to files panel after app search
					break
				}
			}
		} else if m.focusArea == FileTrayFocus && len(m.filteredFiles) > 0 {
			for i, file := range m.fileList {
				if file == m.filteredFiles[0] {
					m.activeFileTab = i
					m.updatePreview(file)
					break
				}
			}
		}
		m.searchQuery = ""
		m.updateFilteredLists()
	case "esc", "ctrl+c":
		m.searchMode = false
		m.searchQuery = ""
		m.updateFilteredLists()
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.updateFilteredLists()
		}
	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.updateFilteredLists()
		}
	}

	return m, nil
}

func (m *Model) updateFilteredLists() {
	if !m.searchMode || m.searchQuery == "" {
		m.filteredApps = m.appList
		m.filteredFiles = m.fileList
		return
	}

	query := strings.ToLower(m.searchQuery)

	m.filteredApps = nil
	for _, app := range m.appList {
		appConfig := m.registry.Apps[app]
		if strings.Contains(strings.ToLower(app), query) ||
			strings.Contains(strings.ToLower(appConfig.Description), query) {
			m.filteredApps = append(m.filteredApps, app)
		}
	}

	m.filteredFiles = nil
	if m.currentApp != "" {
		for _, fileName := range m.fileList {
			fileConfig := m.registry.Apps[m.currentApp].Files[fileName]
			if strings.Contains(strings.ToLower(fileName), query) ||
				strings.Contains(strings.ToLower(fileConfig.Description), query) ||
				strings.Contains(strings.ToLower(fileConfig.Path), query) {
				m.filteredFiles = append(m.filteredFiles, fileName)
			}
		}
	}
}

func (m *Model) GetSelectedApp() string {
	return m.currentApp
}

func (m *Model) GetSelectedFile() string {
	return m.selectedFile
}

func (m *Model) IsQuitting() bool {
	return m.quitting
}

func (m *Model) scrollPreviewToMatch() {
	if len(m.previewMatchIndices) == 0 {
		return
	}
	pos := m.previewMatchIndices[m.previewMatchIndex]
	content := m.previewViewport.View()
	// Count newlines before pos to get line number
	line := 0
	for i := 0; i < pos && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}
	m.previewViewport.GotoTop()
	m.previewViewport.ScrollDown(line)
}
