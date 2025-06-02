package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"hydectl/internal/config"
)

// ViewState represents the current view in the TUI
type ViewState int

const (
	AppSelectionView ViewState = iota
	FileSelectionView
)

// Model represents the TUI model with fzf-like preview
type Model struct {
	registry     *config.ConfigRegistry
	appList      []string
	fileList     []string
	filteredList []string        // For search results
	fileExists   map[string]bool // Track which files exist
	currentApp   string
	cursor       int
	viewState    ViewState
	err          error
	quitting     bool
	selectedFile string
	searchQuery  string // Current search query
	searchMode   bool   // Whether we're in search mode
	windowWidth  int    // Terminal width
	windowHeight int    // Terminal height

	// Viewport for file preview scrolling
	previewViewport viewport.Model
}

// NewModel creates a new TUI model
func NewModel(registry *config.ConfigRegistry) *Model {
	var apps []string
	for appName := range registry.Apps {
		apps = append(apps, appName)
	}
	sort.Strings(apps)

	// Initialize viewport for file preview
	vp := viewport.New(60, 25) // Default size, will be updated on window resize
	vp.YPosition = 0

	return &Model{
		registry:        registry,
		appList:         apps,
		viewState:       AppSelectionView,
		cursor:          0,
		fileExists:      make(map[string]bool),
		windowWidth:     120, // Default width
		windowHeight:    30,  // Default height
		previewViewport: vp,
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		// Update viewport size when window changes
		if m.viewState == FileSelectionView {
			previewWidth := msg.Width/2 - 3
			previewHeight := msg.Height - 10 // Account for headers and footers
			if previewHeight < 10 {
				previewHeight = 10
			}
			m.previewViewport.Width = previewWidth
			m.previewViewport.Height = previewHeight
		}

	case tea.MouseMsg:
		// Handle mouse events for scrolling
		return m.handleMouseEvent(msg)

	case tea.KeyMsg:
		// Handle search mode differently
		if m.searchMode {
			return m.handleSearchMode(msg)
		}

		// Normal navigation mode
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "/":
			// Enter search mode
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
			// Scroll preview up
			if m.viewState == FileSelectionView {
				m.previewViewport.LineUp(10)
			}

		case "pgdn", "ctrl+d":
			// Scroll preview down
			if m.viewState == FileSelectionView {
				m.previewViewport.LineDown(10)
			}

		case "ctrl+k":
			// Scroll preview up by one line
			if m.viewState == FileSelectionView {
				m.previewViewport.LineUp(1)
			}

		case "ctrl+j":
			// Scroll preview down by one line
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

// handleSearchMode handles key presses in search mode
func (m *Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Select from filtered results
		if len(m.filteredList) > 0 && m.cursor < len(m.filteredList) {
			switch m.viewState {
			case AppSelectionView:
				selectedApp := m.filteredList[m.cursor]
				m.currentApp = selectedApp
				appConfig := m.registry.Apps[selectedApp]

				// Exit search mode
				m.searchMode = false
				m.searchQuery = ""

				// If only one file, select it directly and quit
				if len(appConfig.Files) == 1 {
					for fileName := range appConfig.Files {
						m.selectedFile = fileName
						return m, tea.Quit
					}
				}

				// Multiple files, switch to file selection view
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
		// Exit search mode
		m.searchMode = false
		m.searchQuery = ""
		m.cursor = 0
		m.updateFilteredList()

	case "backspace":
		// Remove last character from search query
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
		// Add character to search query
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.cursor = 0
			m.updateFilteredList()
		}
	}
	return m, nil
}

// handleEnter processes the Enter key based on current view state
func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.viewState {
	case AppSelectionView:
		if m.cursor < len(m.appList) {
			selectedApp := m.appList[m.cursor]
			m.currentApp = selectedApp
			appConfig := m.registry.Apps[selectedApp]

			// If only one file, select it directly and quit
			if len(appConfig.Files) == 1 {
				for fileName := range appConfig.Files {
					m.selectedFile = fileName
					return m, tea.Quit
				}
			}

			// Multiple files, switch to file selection view
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

// moveCursor moves the cursor up or down, skipping missing files in file view
func (m *Model) moveCursor(direction int) {
	var displayList []string
	switch m.viewState {
	case AppSelectionView:
		if m.searchMode {
			displayList = m.filteredList
		} else {
			displayList = m.appList
		}
		// For app selection, just move cursor normally
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
		// For file selection, skip missing files
		m.skipMissingFiles(direction)
	}
}

// checkFileExists checks if configuration files exist and updates the fileExists map
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

// canSelectFile returns true if the file can be selected (exists)
func (m *Model) canSelectFile(fileName string) bool {
	exists, found := m.fileExists[fileName]
	return found && exists
}

// skipMissingFiles moves cursor to next/previous available file
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
				m.cursor = 0 // Wrap around
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(displayList) - 1 // Wrap around
			}
		}

		// Check if current file can be selected
		if m.canSelectFile(displayList[m.cursor]) {
			break
		}

		// If we've cycled through all files, break to avoid infinite loop
		if m.cursor == startCursor {
			break
		}
	}
}

// updateFilteredList updates the filtered list based on the current search query
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
			// Search in app name and description
			if strings.Contains(strings.ToLower(app), query) ||
				strings.Contains(strings.ToLower(appConfig.Description), query) {
				m.filteredList = append(m.filteredList, app)
			}
		}
	case FileSelectionView:
		for _, fileName := range m.fileList {
			fileConfig := m.registry.Apps[m.currentApp].Files[fileName]
			// Search in file name, description, and path
			if strings.Contains(strings.ToLower(fileName), query) ||
				strings.Contains(strings.ToLower(fileConfig.Description), query) ||
				strings.Contains(strings.ToLower(fileConfig.Path), query) {
				m.filteredList = append(m.filteredList, fileName)
			}
		}
	}
}

// GetSelectedApp returns the currently selected app
func (m *Model) GetSelectedApp() string {
	return m.currentApp
}

// GetSelectedFile returns the currently selected file
func (m *Model) GetSelectedFile() string {
	return m.selectedFile
}

// IsQuitting returns whether the user wants to quit
func (m *Model) IsQuitting() bool {
	return m.quitting
}

// handleMouseEvent handles mouse events for scrolling and navigation
func (m *Model) handleMouseEvent(msg tea.MouseMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.MouseWheelUp:
		// Scroll up in preview pane if we're in file selection view
		if m.viewState == FileSelectionView {
			m.previewViewport.LineUp(3) // Scroll up 3 lines
		} else if m.viewState == AppSelectionView {
			// Navigate up in application list
			m.moveCursor(-1)
		}

	case tea.MouseWheelDown:
		// Scroll down in preview pane if we're in file selection view
		if m.viewState == FileSelectionView {
			m.previewViewport.LineDown(3) // Scroll down 3 lines
		} else if m.viewState == AppSelectionView {
			// Navigate down in application list
			m.moveCursor(1)
		}

	case tea.MouseLeft:
		// Handle clicks for navigation
		// Calculate which pane was clicked based on mouse position
		totalWidth := m.windowWidth
		if totalWidth < 80 {
			totalWidth = 80
		}
		listWidth := totalWidth / 2

		if msg.X <= listWidth {
			// Click in left pane (list)
			switch m.viewState {
			case AppSelectionView:
				// Calculate which app was clicked
				// Accounting for headers (approximately 6 lines)
				if msg.Y >= 6 && msg.Y < 6+len(m.appList) {
					newCursor := msg.Y - 6
					if newCursor >= 0 && newCursor < len(m.appList) {
						m.cursor = newCursor
					}
				}
			case FileSelectionView:
				// Calculate which file was clicked
				if msg.Y >= 6 && msg.Y < 6+len(m.fileList) {
					newCursor := msg.Y - 6
					if newCursor >= 0 && newCursor < len(m.fileList) {
						m.cursor = newCursor
						// Skip missing files
						m.skipMissingFiles(0)
					}
				}
			}
		} else {
			// Click in right pane (preview)
			if m.viewState == FileSelectionView {
				// Allow clicking in preview to scroll
				// You could add more sophisticated click-to-scroll logic here
			}
		}
	}

	return m, nil
}
