package tui

import (
	"bufio"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

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
}

func NewModel(registry *config.ConfigRegistry) *Model {
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
		return m.handleMouseEvent(msg)

	case tea.KeyMsg:
		if m.searchMode {
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

		case "enter":
			return m.handleEnter()

		case "up", "k":
			m.navigateUp()

		case "down", "j":
			m.navigateDown()

		case "left", "h":
			if m.focusArea == FileTrayFocus {
				m.expandedAppTab = -1
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

		case "space":
			if m.focusArea == AppTabsFocus {
				if m.expandedAppTab == m.activeAppTab {
					m.expandedAppTab = -1
				} else {
					m.expandAppTab(m.activeAppTab)
				}
			}

		case "pgup", "ctrl+u":
			if m.focusArea == PreviewFocus {
				m.previewViewport.LineUp(10)
			}

		case "pgdn", "ctrl+d":
			if m.focusArea == PreviewFocus {
				m.previewViewport.LineDown(10)
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

func (m *Model) navigateUp() {
	switch m.focusArea {
	case AppTabsFocus:
		if m.activeAppTab > 0 {
			m.activeAppTab--
		}
	case FileTrayFocus:
		if m.activeFileTab > 0 {
			m.activeFileTab--
		}
	case PreviewFocus:
		m.previewViewport.LineUp(1)
	}
}

func (m *Model) navigateDown() {
	switch m.focusArea {
	case AppTabsFocus:
		if m.activeAppTab < len(m.appList)-1 {
			m.activeAppTab++
		}
	case FileTrayFocus:
		if m.activeFileTab < len(m.fileList)-1 {
			m.activeFileTab++
		}
	case PreviewFocus:
		m.previewViewport.LineDown(1)
	}
}

func (m *Model) expandAppTab(appIndex int) {
	if appIndex >= 0 && appIndex < len(m.appList) {
		m.expandedAppTab = appIndex
		m.currentApp = m.appList[appIndex]
		m.loadFileList()
		m.focusArea = FileTrayFocus
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

func (m *Model) updatePreview(fileName string) {
	if m.currentApp == "" || fileName == "" {
		return
	}

	appConfig := m.registry.Apps[m.currentApp]
	fileConfig, exists := appConfig.Files[fileName]
	if !exists {
		return
	}

	var lines []string

	headerText := fileName
	lines = append(lines, ColorBrightCyan+headerText+ColorReset)
	lines = append(lines, ColorBrightBlack+strings.Repeat("─", len(headerText))+ColorReset)
	lines = append(lines, "")

	lines = append(lines, ColorBrightYellow+"Description:"+ColorReset+" "+fileConfig.Description)
	lines = append(lines, ColorBrightYellow+"Path:"+ColorReset+" "+fileConfig.Path)
	lines = append(lines, "")

	if fileConfig.FileExists() {
		lines = append(lines, ColorBrightGreen+"✓ File exists"+ColorReset)
		lines = append(lines, "")
		lines = append(lines, ColorBrightYellow+"Content:"+ColorReset)
		lines = append(lines, ColorBrightBlack+strings.Repeat("─", 40)+ColorReset)

		expandedPath := config.ExpandPath(fileConfig.Path)
		content, _ := m.readFileContent(expandedPath)
		lines = append(lines, content...)
	} else {
		lines = append(lines, ColorBrightRed+"❌ File does not exist"+ColorReset)
		lines = append(lines, "")
		lines = append(lines, ColorBrightBlack+"File will be created when selected"+ColorReset)
	}

	m.previewViewport.SetContent(strings.Join(lines, "\n"))
}

func (m *Model) readFileContent(filePath string) ([]string, error) {

	content, _ := m.readFilePreviewWithScroll(filePath)
	return content, nil
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

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.focusArea {
	case AppTabsFocus:
		if m.expandedAppTab == m.activeAppTab {

			m.expandedAppTab = -1
		} else {

			m.expandAppTab(m.activeAppTab)
		}
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

func (m *Model) handleMouseEvent(msg tea.MouseMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.MouseWheelUp:
		if m.focusArea == PreviewFocus {
			m.previewViewport.LineUp(3)
		}

	case tea.MouseWheelDown:
		if m.focusArea == PreviewFocus {
			m.previewViewport.LineDown(3)
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
