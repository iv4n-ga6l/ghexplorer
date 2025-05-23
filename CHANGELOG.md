# Changelog

All notable changes to the GitHub Profile Explorer project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Enhanced loading states with spinner animations
- Status messages for better user feedback
- Keyboard shortcuts for improved navigation:
  - `h` - Show help message
  - `r` - Refresh repositories
  - `Ctrl+R` - Force refresh profile
- Loading view with animated spinner
- Better error handling and user feedback
- Status messages in footer showing operation results
- Comprehensive documentation in ENHANCED_README.md

### Changed
- Improved TUI responsiveness with loading indicators
- Enhanced user experience with real-time feedback
- Better navigation flow between different views
- Optimized API call handling with loading states

### Fixed
- Go version compatibility issues with min/max functions
- Added helper functions for older Go versions
- Improved error handling throughout the application

### Technical Improvements
- Added `Min` and `Max` utility functions in helper package
- Enhanced model structure with loading state management
- Improved spinner integration with Bubble Tea framework
- Better separation of concerns in view rendering

## [1.0.0] - Initial Release

### Added
- GitHub Profile Explorer TUI application
- Profile viewing with user information
- Repository browsing and exploration
- File content viewing with syntax highlighting
- Repository search functionality
- Command-line interface with multiple commands:
  - `explore` - Interactive TUI mode
  - `repo` - Repository information
  - `search` - Repository search
- JSON and text output formats
- Clipboard integration for copying content
- Responsive design with lipgloss styling
- Comprehensive test coverage
- Cross-platform compatibility

### Features
- Interactive Terminal User Interface (TUI)
- GitHub API integration
- Real-time data fetching
- Keyboard navigation
- Text selection and copying
- Pagination for large datasets
- Error handling and user feedback
- Configuration management
- Modern CLI design with Cobra framework