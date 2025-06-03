package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"hydectl/internal/config"
)

type ViewState int

const (
	AppSelectionView ViewState = iota
	FileSelectionView
)

type Model struct {
	registry     *config.ConfigRegistry
	appList      []string
	fileList     []string
	filteredList []string
	fileExists   map[string]bool
	currentApp   string
	cursor       int
	viewState    ViewState
	err          error
	quitting     bool
	selectedFile string
	searchQuery  string
	searchMode   bool
	windowWidth  int
	windowHeight int

	previewViewport viewport.Model
}

func NewModel(registry *config.ConfigRegistry) *Model {
	var apps []string
	for appName := range registry.Apps {
		apps = append(apps, appName)
	}
	sort.Strings(apps)

	vp := viewport.New(60, 25)
	vp.YPosition = 0

	return &Model{
		registry:        registry,
		appList:         apps,
		viewState:       AppSelectionView,
		cursor:          0,
		fileExists:      make(map[string]bool),
		windowWidth:     120,
		windowHeight:    30,
		previewViewport: vp,
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

		if m.viewState == FileSelectionView {
			previewWidth := msg.Width/2 - 3
			previewHeight := msg.Height - 10
			if previewHeight < 10 {
				previewHeight = 10
			}
			m.previewViewport.Width = previewWidth
			m.previewViewport.Height = previewHeight
		}

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
			m.cursor = 0
			m.updateFilteredList()

		case "enter":
			return m.handleEnter()

		case "up", "k":
			m.moveCursor(-1)

		case "down", "j":
			m.moveCursor(1)

		case "pgup", "ctrl+u":

			if m.viewState == FileSelectionView {
				m.previewViewport.LineUp(10)
			}

		case "pgdn", "ctrl+d":

			if m.viewState == FileSelectionView {
				m.previewViewport.LineDown(10)
			}

		case "ctrl+k":

			if m.viewState == FileSelectionView {
				m.previewViewport.LineUp(1)
			}

		case "ctrl+j":

			if m.viewState == FileSelectionView {
				m.previewViewport.LineDown(1)
			}

		case "backspace", "left", "h":
			if m.viewState == FileSelectionView {
				m.viewState = AppSelectionView
				m.cursor = 0
				m.currentApp = ""
				m.fileList = nil
			}
		}
	}

	return m, nil
}

func (m *Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":

		if len(m.filteredList) > 0 && m.cursor < len(m.filteredList) {
			switch m.viewState {
			case AppSelectionView:
				selectedApp := m.filteredList[m.cursor]
				m.currentApp = selectedApp
				appConfig := m.registry.Apps[selectedApp]

				m.searchMode = false
				m.searchQuery = ""

				if len(appConfig.Files) == 1 {
					for fileName := range appConfig.Files {
						m.selectedFile = fileName
						return m, tea.Quit
					}
				}

				var files []string
				for fileName := range appConfig.Files {
					files = append(files, fileName)
				}
				sort.Strings(files)

				m.fileList = files
				m.viewState = FileSelectionView
				m.cursor = 0
				m.checkFileExists()

			case FileSelectionView:
				if m.canSelectFile(m.filteredList[m.cursor]) {
					m.selectedFile = m.filteredList[m.cursor]
					return m, tea.Quit
				}
			}
		}

	case "esc", "ctrl+c":

		m.searchMode = false
		m.searchQuery = ""
		m.cursor = 0
		m.updateFilteredList()

	case "backspace":

		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.cursor = 0
			m.updateFilteredList()
		}

	case "up", "ctrl+k":
		m.moveCursor(-1)

	case "down", "ctrl+j":
		m.moveCursor(1)

	default:

		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.cursor = 0
			m.updateFilteredList()
		}
	}
	return m, nil
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.viewState {
	case AppSelectionView:
		if m.cursor < len(m.appList) {
			selectedApp := m.appList[m.cursor]
			m.currentApp = selectedApp
			appConfig := m.registry.Apps[selectedApp]

			if len(appConfig.Files) == 1 {
				for fileName := range appConfig.Files {
					m.selectedFile = fileName
					return m, tea.Quit
				}
			}

			var files []string
			for fileName := range appConfig.Files {
				files = append(files, fileName)
			}
			sort.Strings(files)

			m.fileList = files
			m.viewState = FileSelectionView
			m.cursor = 0
			m.checkFileExists()
		}

	case FileSelectionView:
		if m.cursor < len(m.fileList) && m.canSelectFile(m.fileList[m.cursor]) {
			m.selectedFile = m.fileList[m.cursor]
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *Model) moveCursor(direction int) {
	var displayList []string
	switch m.viewState {
	case AppSelectionView:
		if m.searchMode {
			displayList = m.filteredList
		} else {
			displayList = m.appList
		}

		if direction > 0 {
			if m.cursor < len(displayList)-1 {
				m.cursor++
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
			}
		}

	case FileSelectionView:

		m.skipMissingFiles(direction)
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

func (m *Model) canSelectFile(fileName string) bool {
	exists, found := m.fileExists[fileName]
	return found && exists
}

func (m *Model) skipMissingFiles(direction int) {
	var displayList []string
	if m.searchMode {
		displayList = m.filteredList
	} else {
		displayList = m.fileList
	}

	if len(displayList) == 0 {
		return
	}

	startCursor := m.cursor
	for {
		if direction > 0 {
			if m.cursor < len(displayList)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(displayList) - 1
			}
		}

		if m.canSelectFile(displayList[m.cursor]) {
			break
		}

		if m.cursor == startCursor {
			break
		}
	}
}

func (m *Model) updateFilteredList() {
	if !m.searchMode {
		return
	}

	query := strings.ToLower(m.searchQuery)
	m.filteredList = nil

	switch m.viewState {
	case AppSelectionView:
		for _, app := range m.appList {
			appConfig := m.registry.Apps[app]

			if strings.Contains(strings.ToLower(app), query) ||
				strings.Contains(strings.ToLower(appConfig.Description), query) {
				m.filteredList = append(m.filteredList, app)
			}
		}
	case FileSelectionView:
		for _, fileName := range m.fileList {
			fileConfig := m.registry.Apps[m.currentApp].Files[fileName]

			if strings.Contains(strings.ToLower(fileName), query) ||
				strings.Contains(strings.ToLower(fileConfig.Description), query) ||
				strings.Contains(strings.ToLower(fileConfig.Path), query) {
				m.filteredList = append(m.filteredList, fileName)
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

func (m *Model) handleMouseEvent(msg tea.MouseMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.MouseWheelUp:

		if m.viewState == FileSelectionView {
			m.previewViewport.LineUp(3)
		} else if m.viewState == AppSelectionView {

			m.moveCursor(-1)
		}

	case tea.MouseWheelDown:

		if m.viewState == FileSelectionView {
			m.previewViewport.LineDown(3)
		} else if m.viewState == AppSelectionView {

			m.moveCursor(1)
		}

	case tea.MouseLeft:

		totalWidth := m.windowWidth
		if totalWidth < 80 {
			totalWidth = 80
		}
		listWidth := totalWidth / 2

		if msg.X <= listWidth {

			switch m.viewState {
			case AppSelectionView:

				if msg.Y >= 6 && msg.Y < 6+len(m.appList) {
					newCursor := msg.Y - 6
					if newCursor >= 0 && newCursor < len(m.appList) {
						m.cursor = newCursor
					}
				}
			case FileSelectionView:

				if msg.Y >= 6 && msg.Y < 6+len(m.fileList) {
					newCursor := msg.Y - 6
					if newCursor >= 0 && newCursor < len(m.fileList) {
						m.cursor = newCursor

						m.skipMissingFiles(0)
					}
				}
			}
		} else {

			if m.viewState == FileSelectionView {

			}
		}
	}

	return m, nil
}
