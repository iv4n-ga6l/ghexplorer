package model

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ghexplorer/config"
	"ghexplorer/github_api"
	"ghexplorer/helper"
)

// Model is the base model
type Model struct {
	githubID     string
	inputting    bool
	profile      *github_api.GitHubProfile
	repositories []*github_api.Repository
	currentView  string
	cursor       int
	selected     map[string]string
	fileContents []*github_api.FileInfo
	fileContent  string
	searchQuery  string
	errorMessage string
	selectMode   bool
	selectStart  int
	selectEnd    int
	textInput    textinput.Model
	viewport     viewport.Model
	spinner      spinner.Model
	tabs         []string
	activeTab    int
	loading      bool
	loadingMsg   string
	statusMsg    string
}

// InitialModel initialModel initialize the model
func InitialModel(initialGithubID string) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter GitHub profile ID"
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = config.SpinnerStyle

	vp := viewport.New(80, 20)
	vp.YPosition = config.HeaderHeight

	m := Model{
		githubID:    initialGithubID,
		inputting:   initialGithubID == "",
		currentView: "input",
		selected:    make(map[string]string),
		textInput:   ti,
		spinner:     s,
		tabs:        []string{"Overview", "Repositories"},
		activeTab:   0,
		viewport:    vp,
		loading:     false,
		loadingMsg:  "",
		statusMsg:   "",
	}

	// If initial GitHub ID is provided, set it in the text input
	if initialGithubID != "" {
		ti.SetValue(initialGithubID)
		ti.Blur()
	}

	return m
}

// Init the model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles the CLI view interaction updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "tab":
			if m.currentView == "profile" || m.currentView == "repositories" {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				switch m.activeTab {
				case 0:
					m.currentView = "profile"
				case 1:
					m.currentView = "repositories"
				}
			}
		case "r":
			if m.currentView == "repositories" || m.currentView == "profile" {
				m.loading = true
				m.loadingMsg = "Refreshing repositories..."
				m.currentView = "loading"
				return m, tea.Batch(m.spinner.Tick, m.fetchRepositories)
			}
		case "h":
			// Show help
			if m.currentView != "input" && m.currentView != "loading" {
				m.statusMsg = "Help: ↑/↓ navigate, Enter select, Tab switch, / search, r refresh, q quit"
			}
		case "ctrl+r":
			// Force refresh profile
			if m.githubID != "" && m.currentView != "input" && m.currentView != "loading" {
				m.loading = true
				m.loadingMsg = "Refreshing profile..."
				m.currentView = "loading"
				return m, tea.Batch(m.spinner.Tick, m.fetchProfile)
			}
		case "enter":
			switch m.currentView {
			case "input":
				m.githubID = m.textInput.Value()
				m.inputting = false
				m.currentView = "loading"
				m.loading = true
				m.loadingMsg = "Fetching profile..."
				return m, tea.Batch(m.spinner.Tick, m.fetchProfile)
			case "repositories":
				if m.cursor < len(m.repositories) {
					m.selected["repository"] = m.repositories[m.cursor].Name
					m.currentView = "loading"
					m.cursor = 0
					m.selected["path"] = ""
					m.loading = true
					m.loadingMsg = "Loading repository contents..."
					return m, tea.Batch(m.spinner.Tick, m.fetchRepositoryContents)
				}
			case "files":
				if m.cursor < len(m.fileContents) {
					if m.fileContents[m.cursor].Type == "file" {
						m.selected["file"] = m.fileContents[m.cursor].Name
						m.currentView = "loading"
						m.loading = true
						m.loadingMsg = "Loading file content..."
						return m, tea.Batch(m.spinner.Tick, m.fetchFileContent)
					} else {
						m.selected["path"] += "/" + m.fileContents[m.cursor].Name
						m.currentView = "loading"
						m.loading = true
						m.loadingMsg = "Loading directory..."
						return m, tea.Batch(m.spinner.Tick, m.fetchRepositoryContents)
					}
				}
			case "search":
				m.currentView = "loading"
				m.loading = true
				m.loadingMsg = "Searching repositories..."
				return m, tea.Batch(m.spinner.Tick, m.searchRepositories)
			}
		case "backspace":
			if m.inputting && len(m.githubID) > 0 {
				m.githubID = m.githubID[:len(m.githubID)-1]
			} else if m.currentView == "search" && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			}
		case "esc":
			if m.selectMode {
				m.selectMode = false
				m.selectStart = 0
				m.selectEnd = 0
			} else {
				switch m.currentView {
				case "repositories":
					m.currentView = "profile"
				case "files":
					if m.selected["path"] == "" {
						m.currentView = "repositories"
						m.cursor = 0
					} else {
						paths := strings.Split(m.selected["path"], "/")
						m.selected["path"] = strings.Join(paths[:len(paths)-1], "/")
						return m, m.fetchRepositoryContents
					}
				case "fileContent":
					m.currentView = "files"
					m.selectMode = false
					m.selectStart = 0
					m.selectEnd = 0
				case "search":
					m.currentView = "repositories"
				}
			}
		case "up", "down", "pgup", "pgdown":
			if m.currentView == "fileContent" {
				if m.selectMode {
					switch msg.String() {
					case "up":
						m.selectEnd = helper.Max(0, m.selectEnd-m.viewport.Width)
					case "down":
						m.selectEnd = helper.Min(len(m.fileContent), m.selectEnd+m.viewport.Width)
					case "left":
						m.selectEnd = helper.Max(0, m.selectEnd-1)
					case "right":
						m.selectEnd = helper.Min(len(m.fileContent), m.selectEnd+1)
					}
				}
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			} else {
				switch msg.String() {
				case "up":
					if m.cursor > 0 {
						m.cursor--
					}
				case "down":
					switch m.currentView {
					case "repositories":
						if m.cursor < len(m.repositories)-1 {
							m.cursor++
						}
					case "files":
						if m.cursor < len(m.fileContents)-1 {
							m.cursor++
						}
					}
				}
			}
		case "ctrl+a":
			if m.currentView == "fileContent" {
				m.selectMode = true
				m.selectStart = 0
				m.selectEnd = len(m.fileContent)
			}
		case "ctrl+c":
			if m.selectMode && m.currentView == "fileContent" {
				selectedText := m.fileContent[m.selectStart:m.selectEnd]
				clipboard.WriteAll(selectedText)
				m.selectMode = false
				m.selectStart = 0
				m.selectEnd = 0
				return m, nil
			}
		case "ctrl+d":
			if m.selectMode && m.currentView == "fileContent" {
				m.selectMode = false
				m.selectStart = 0
				m.selectEnd = len(m.fileContent)
				return m, nil
			}
		case "/":
			if m.currentView == "repositories" {
				m.currentView = "search"
				m.searchQuery = ""
			}
		default:
			if m.inputting {
				m.githubID += msg.String()
			} else if m.currentView == "search" {
				m.searchQuery += msg.String()
			}
		}
	case tea.WindowSizeMsg:
		width, height := m.calculateViewportDimensions(msg)
		m.viewport = viewport.New(width, height)
		m.viewport.YPosition = config.HeaderHeight
		m.viewport.HighPerformanceRendering = config.UseHighPerformanceRenderer
		if m.currentView == "fileContent" {
			m.viewport.SetContent(m.fileContent)
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case github_api.GitHubProfile:
		m.profile = &msg
		m.loading = true
		m.loadingMsg = "Loading repositories..."
		return m, tea.Batch(m.spinner.Tick, m.fetchRepositories)
	case []*github_api.Repository:
		m.repositories = msg
		m.currentView = "repositories"
		m.cursor = 0
		m.loading = false
		m.statusMsg = fmt.Sprintf("Loaded %d repositories", len(msg))
	case []*github_api.FileInfo:
		m.fileContents = msg
		m.currentView = "files"
		m.cursor = 0
		m.loading = false
		m.statusMsg = fmt.Sprintf("Loaded %d items", len(msg))
	case string:
		m.fileContent = msg
		m.currentView = "fileContent"
		m.loading = false
		m.statusMsg = "File loaded successfully"
		if m.viewport.Height > 0 {
			m.viewport.SetContent(m.fileContent)
			m.viewport.GotoTop()
		}
	case error:
		m.currentView = "error"
		m.errorMessage = msg.Error()
		m.loading = false
		return m, nil
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// fetchProfile handles the profile fetching
func (m Model) fetchProfile() tea.Msg {
	profile, err := github_api.FetchGitHubProfile(m.githubID)
	if err != nil {
		return err
	}

	return *profile
}

// fetchRepositories handles the profile repositories fetching
func (m Model) fetchRepositories() tea.Msg {
	repos, err := github_api.FetchRepositories(m.profile.Login)
	if err != nil {
		return err
	}
	return repos
}

// fetchRepositoryContents handles the profile repository contents fetching
func (m Model) fetchRepositoryContents() tea.Msg {
	contents, err := github_api.FetchRepositoryContents(m.profile.Login, m.selected["repository"], m.selected["path"])
	if err != nil {
		return err
	}
	return contents
}

// fetchFileContent handles the profile repository file content fetching
func (m Model) fetchFileContent() tea.Msg {
	content, err := github_api.FetchFileContent(m.profile.Login, m.selected["repository"], m.selected["path"]+"/"+m.selected["file"])
	if err != nil {
		return err
	}
	return content
}

// searchRepositories handles the profile repositories search performing
func (m Model) searchRepositories() tea.Msg {
	repos, err := github_api.SearchRepositories(m.profile.Login, m.searchQuery)
	if err != nil {
		return err
	}
	return repos
}

// View handles the CLI global view
func (m Model) View() string {
	switch m.currentView {
	case "input":
		return m.inputView()
	case "loading":
		return m.loadingView()
	case "profile", "repositories":
		return m.tabView()
	case "files":
		return m.filesView()
	case "fileContent":
		return m.fileContentView()
	case "search":
		return m.searchView()
	case "error":
		return m.errorView()
	default:
		return config.DocStyle.Render(
			config.CardStyle.Render(
				lipgloss.JoinHorizontal(
					lipgloss.Center,
					m.spinner.View(),
					" Loading...",
				),
			),
		)
	}
}

// calculateViewportDimensions calculates the available viewport dimensions
func (m Model) calculateViewportDimensions(msg tea.WindowSizeMsg) (width, height int) {
	// Subtract padding and borders from total width
	width = msg.Width - 4 // 2 for left padding + 2 for right padding

	// Subtract header, footer, and padding from total height
	height = msg.Height - config.HeaderHeight - config.FooterHeight - 2 // 2 for top/bottom padding

	return width, height
}

// getPaginationInfo returns pagination details for the current view
func (m Model) getPaginationInfo() (currentPage, totalPages, startIdx, endIdx int) {
	var totalItems int

	switch m.currentView {
	case "repositories":
		totalItems = len(m.repositories)
	case "files":
		totalItems = len(m.fileContents)
	default:
		return 1, 1, 0, 0
	}

	currentPage = (m.cursor / config.ItemsPerPage) + 1
	totalPages = (totalItems + config.ItemsPerPage - 1) / config.ItemsPerPage

	startIdx = (currentPage - 1) * config.ItemsPerPage
	endIdx = helper.Min(startIdx+config.ItemsPerPage, totalItems)

	return currentPage, totalPages, startIdx, endIdx
}

// renderPagination renders the pagination information
func renderPagination(current, total int) string {
	if total <= 1 {
		return ""
	}
	return config.PaginationInfoStyle.Render(fmt.Sprintf(config.PaginationStyle, current, total))
}

// tabView handles the CLI tab view
func (m Model) tabView() string {
	doc := strings.Builder{}

	// Render tabs
	renderedTabs := []string{}
	for i, t := range m.tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, config.ActiveTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, config.TabStyle.Render(t))
		}
	}
	doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...))
	doc.WriteString("\n\n")

	// Render content based on active tab
	switch m.activeTab {
	case 0: // Overview
		doc.WriteString(m.overviewView())
	case 1: // Repositories
		doc.WriteString(m.repositoriesView())
	}

	return config.DocStyle.Render(doc.String())
}

// overviewView handles the CLI overview view - simplified without README
func (m Model) overviewView() string {
	if m.profile == nil {
		return config.CardStyle.Render("Loading profile...")
	}

	// Profile information section
	profileInfo := lipgloss.JoinVertical(
		lipgloss.Left,
		config.HeaderStyle.Render("Profile Information"),
		lipgloss.JoinHorizontal(lipgloss.Left, config.LabelStyle.Render("Name:"), config.ValueStyle.Render(helper.StringOrNA(m.profile.Name))),
		lipgloss.JoinHorizontal(lipgloss.Left, config.LabelStyle.Render("Username:"), config.ValueStyle.Render(helper.StringOrNA(m.profile.Login))),
		lipgloss.JoinHorizontal(lipgloss.Left, config.LabelStyle.Render("Bio:"), config.ValueStyle.Render(helper.StringOrNA(m.profile.Description))),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			config.LabelStyle.Render("Stats:"),
			config.ValueStyle.Render(
				fmt.Sprintf("👥 %d followers • %d following", m.profile.Followers, m.profile.Following),
			),
		),
	)

	return config.ProfileCardStyle.Render(profileInfo)
}

// repositoriesView handles the CLI repositories view
func (m Model) repositoriesView() string {
	var content strings.Builder

	currentPage, totalPages, startIdx, endIdx := m.getPaginationInfo()

	content.WriteString(config.HeaderStyle.Render("Repositories"))
	content.WriteString("\n\n")

	// Display only the repositories for the current page
	visibleRepos := m.repositories[startIdx:endIdx]
	for i, repo := range visibleRepos {
		cursor := " "
		if startIdx+i == m.cursor {
			cursor = ">"
		}

		repoCard := lipgloss.JoinVertical(
			lipgloss.Left,
			config.RepositoryStyle.Render(helper.StringOrNA(repo.Name)),
			config.ValueStyle.Render(helper.StringOrNA(repo.Description)),
		)

		if startIdx+i == m.cursor {
			repoCard = config.SelectedStyle.Render(repoCard)
		} else {
			repoCard = config.CardStyle.Render(repoCard)
		}

		content.WriteString(fmt.Sprintf("%s %s\n", cursor, repoCard))
	}

	// Add pagination info
	content.WriteString("\n")
	content.WriteString(renderPagination(currentPage, totalPages))

	footer := config.FooterStyle.Render("\nPress Enter to view files • '/' to search • Tab to switch tabs • ←/→ to change pages")

	content.WriteString(footer)

	return content.String()
}

// filesView handles the CLI files view
func (m Model) filesView() string {
	var content strings.Builder

	currentPage, totalPages, startIdx, endIdx := m.getPaginationInfo()

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		config.HeaderStyle.Render(fmt.Sprintf("Repository: %s", helper.StringOrNA(m.selected["repository"]))),
		config.ValueStyle.Render(fmt.Sprintf("Path: %s", helper.StringOrNA(m.selected["path"]))),
	)

	content.WriteString(config.CardStyle.Render(header))
	content.WriteString("\n\n")

	// Display only the files for the current page
	visibleFiles := m.fileContents[startIdx:endIdx]
	for i, file := range visibleFiles {
		cursor := " "
		if startIdx+i == m.cursor {
			cursor = ">"
		}

		var fileNameStyle lipgloss.Style
		var icon string
		if file.Type == "dir" {
			fileNameStyle = config.FolderStyle
			icon = "📁"
		} else {
			fileNameStyle = config.FileStyle
			icon = "📄"
		}

		fileCard := lipgloss.JoinHorizontal(
			lipgloss.Left,
			icon,
			" ",
			fileNameStyle.Render(helper.StringOrNA(file.Name)),
		)

		if startIdx+i == m.cursor {
			fileCard = config.SelectedStyle.Render(fileCard)
		} else {
			fileCard = config.CardStyle.Render(fileCard)
		}

		content.WriteString(fmt.Sprintf("%s %s\n", cursor, fileCard))
	}

	// Add pagination info
	content.WriteString("\n")
	content.WriteString(renderPagination(currentPage, totalPages))

	footer := config.FooterStyle.Render("\nPress Enter to view content • Esc to go back • ←/→ to change pages")

	content.WriteString(footer)

	return content.String()
}

// fileContentView handles the CLI fileContent view
func (m Model) fileContentView() string {
	header := config.CardStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			config.HeaderStyle.Render("File Content"),
			config.ValueStyle.Render(helper.StringOrNA(m.selected["file"])),
		),
	)

	footer := config.FooterStyle.Render("\nPress Esc to go back • Ctrl+A to select all • Ctrl+C to copy • Ctrl+D to deselect • ↑/↓ to scroll")

	var styledContent string
	if m.selectMode {
		before := m.fileContent[:m.selectStart]
		selected := config.SelectedStyle.Render(m.fileContent[m.selectStart:m.selectEnd])
		after := m.fileContent[m.selectEnd:]
		styledContent = before + selected + after
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"\n",
			config.CardStyle.Render(styledContent),
			footer,
		)
	} else {
		styledContent = m.fileContent
	}

	// Set the viewport content if it hasn't been set
	if m.viewport.Height == 0 {
		m.viewport.Height = m.viewport.Height - config.HeaderHeight - config.FooterHeight - 4 // Adjust for header, footer, and padding
		m.viewport.Width = m.viewport.Width - 4                                               // Adjust for left and right padding
		m.viewport.SetContent(styledContent)
	}

	// Render the viewport
	content := m.viewport.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"\n",
		config.CardStyle.Render(content),
		footer,
	)
}

// searchView handles the CLI search view
func (m Model) searchView() string {
	return config.DocStyle.Render(
		config.CardStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				config.HeaderStyle.Render("Search Repositories"),
				"",
				config.ValueStyle.Render(m.searchQuery),
				"",
				config.FooterStyle.Render("Press Enter to search • Esc to cancel"),
			),
		),
	)
}

// inputView handles the CLI input view
func (m Model) inputView() string {
	return config.DocStyle.Render(
		config.CardStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				config.HeaderStyle.Render("GitHub Profile Explorer"),
				"",
				"Enter GitHub username:",
				m.textInput.View(),
				"",
				config.FooterStyle.Render("Press Enter to continue • q to quit"),
			),
		),
	)
}

// loadingView handles the CLI loading view
func (m Model) loadingView() string {
	return config.DocStyle.Render(
		config.CardStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				config.HeaderStyle.Render("GitHub Profile Explorer"),
				"",
				lipgloss.JoinHorizontal(
					lipgloss.Center,
					m.spinner.View(),
					" "+m.loadingMsg,
				),
				"",
				config.FooterStyle.Render("Please wait... • q to quit"),
			),
		),
	)
}

// errorView handles the CLI error view
func (m Model) errorView() string {
	return config.DocStyle.Render(
		config.CardStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				config.ErrorStyle.Render("Error Occurred"),
				"",
				config.ValueStyle.Render(m.errorMessage),
				"",
				config.FooterStyle.Render("Press 'q' key to quit"),
			),
		),
	)
}

func (m Model) renderFooter() string {
	var footerText string
	switch m.currentView {
	case "repositories":
		footerText = "↑/↓: navigate • Enter: open repo • Tab: switch tabs • /: search • q: quit"
	case "files":
		footerText = "↑/↓: navigate • Enter: open • Backspace: back • q: quit"
	case "fileContent":
		footerText = "↑/↓/PgUp/PgDn: scroll • Ctrl+A: select all • Ctrl+C: copy • Ctrl+D: copy selection • Backspace: back • q: quit"
	case "search":
		footerText = "Type to search • Enter: search • Esc: back • q: quit"
	default:
		footerText = "q: quit"
	}

	// Add status message if available
	if m.statusMsg != "" {
		footerText = m.statusMsg + " • " + footerText
	}

	return config.FooterStyle.Render(footerText)
}
