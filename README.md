# ghexplorer

ghexplorer is a terminal-based application written in Go that allows users to interactively explore GitHub profiles, repositories, and file contents. This tool provides a user-friendly interface to navigate through GitHub Profile without leaving your terminal.

https://github.com/user-attachments/assets/aa417c9e-3b3d-4ad1-a3e8-ca4991580a25

![Diagram](diagram.png)

## Features

- **Profile Viewing**: Enter a GitHub username to view basic profile information.
- **Repository Listing**: Browse through a user's repositories with descriptions.
- **File Navigation**: Explore repository contents, including folders and files.
- **File Content Display**: View the contents of files directly in the terminal.
- **Repository Search**: Search for specific repositories within a user's profile.
- **Interactive Navigation**: Use keyboard shortcuts to navigate through different views.
- **Color-Coded Display**: Repositories, folders, and files are color-coded for easy identification.
- **Scrollable File Content**: Navigate through long file contents using scroll functionality.
- **Text Selection and Copying**: Select and copy file contents to your clipboard.

## Prerequisites

Before you begin, ensure you have the following installed:
- Go (version 1.16 or later)
- Git

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/IvanGael/ghexplorer.git
   cd ghexplorer
   ```

2. Install the required dependencies:
   ```
   go get github.com/charmbracelet/bubbletea
   go get github.com/charmbracelet/lipgloss
   go get github.com/atotto/clipboard
   go get -u github.com/spf13/cobra
   go get "github.com/stretchr/testify/assert"
   ```

3. Build the application:
   ```
   go build .
   ```

## Usage

1. Run the application:
- Start TUI with empty input
   ```
   ghexplorer explore
   ```
- Start TUI with pre-filled username
   ```
   ghexplorer explore USERNAME
   ```

2. Repository information:
- Get repo info in text format
   ```
   ghexplorer repo USERNAME REPOSITORY_NAME
   ```
- Get repo info in JSON format
   ```
   ghexplorer repo USERNAME REPOSITORY_NAME -f json
   ```
- Save repo info to file
   ```
   ghexplorer repo USERNAME REPOSITORY_NAME -o repo.txt
   ```

3. Search repositories:
- Search repos in text format
   ```
   ghexplorer search USERNAME REPO_SEARCH
   ```
- Search repos in JSON format
   ```
   ghexplorer search USERNAME REPO_SEARCH -f json
   ```
- Save search results to file
   ```
   ghexplorer search USERNAME REPO_SEARCH -o search.txt
   ```

4. Use the following keyboard shortcuts to navigate:
   - Arrow keys: Move cursor / Scroll file contents
   - Enter: Select / Open
   - Esc: Go back / Exit selection mode
   - '/': Enter search mode (when viewing repositories)
   - Ctrl+A: Select all (in file view)
   - Ctrl+C: Copy selected text (in file view)
   - Ctrl+D: Deselect all (in file view)
   - PgUp/PgDown: Scroll file contents quickly
   - 'q': Quit the application

## Customization

You can make any customization regarding styles or colors used in the application by modifying the `config.go` file:

```go
var (
	repositoryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	folderStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	fileStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	selectedStyle   = lipgloss.NewStyle().Background(lipgloss.Color("25"))
)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) for terminal styling
- [Clipboard](https://github.com/atotto/clipboard) for clipboard functionality

## Disclaimer

This application uses the GitHub API without authentication, which has rate limits. For a production application, consider implementing proper authentication using GitHub tokens to increase the rate limits and access private repositories if needed.