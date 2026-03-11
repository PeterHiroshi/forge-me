# Changelog

All notable changes to cfmon will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-03-11

### Added
- `cfmon tail` command for real-time log streaming from Cloudflare Workers and Containers
- WebSocket-based Tail API integration (`internal/api/tail.go`)
- Tail event types and JSON parsing (`internal/tail/types.go`)
- Output formatter with pretty (colored), JSON, and compact modes (`internal/tail/formatter.go`)
- WebSocket engine with auto-reconnect and client-side filtering (`internal/tail/engine.go`)
- 12+ CLI flags for advanced filtering: `--status`, `--method`, `--search`, `--ip`, `--header`, `--sample-rate`, `--max-events`, `--since`, `--no-color`, `--include-logs`, `--include-exceptions`
- `doRequestWithBody` method on API client for POST requests with JSON body
- Comprehensive test coverage (94.8% on internal/tail package)

### Added
- Makefile with comprehensive build targets for easier development
- .goreleaser.yml for automated cross-platform releases
- Enhanced install.sh script with OS/arch detection and colorful output
- CONTRIBUTING.md with detailed contribution guidelines
- Rich help command with examples and better formatting
- Global `--verbose` flag for debug output
- Global `--timeout` flag for API request timeout customization
- `cfmon doctor` command for system health checks
- `cfmon config show` command to display current configuration
- `cfmon config path` command to show config file location
- `--sort` flag for containers/workers commands (by name, cpu, memory, requests)
- `--limit` flag for containers/workers commands to limit results
- `--filter` flag for basic name filtering with substring matching
- User-friendly error messages with suggested fixes
- Comprehensive test coverage for all new features
- Integration tests for end-to-end testing

### Changed
- Improved error handling with actionable suggestions
- Enhanced CLI output formatting for better readability
- Updated README with comprehensive examples and quick reference

### Fixed
- Various bug fixes and performance improvements

## [v0.1.0] - 2024-03-05

### Added
- Initial release of cfmon
- Basic CLI structure using Cobra
- Cloudflare API client implementation
- Container management commands:
  - `cfmon containers list` - List all containers
  - `cfmon containers status <id>` - Get container status
- Worker management commands:
  - `cfmon workers list` - List all workers
  - `cfmon workers status <name>` - Get worker status
- Authentication:
  - `cfmon login <token>` - Store API token securely
- Configuration management with config file support
- Output formats:
  - JSON output with `--json` flag
  - Formatted table output (default)
- Shell completion support:
  - Bash completion
  - Zsh completion
  - Fish completion
  - PowerShell completion
- Installation scripts:
  - Bash installer for Linux/macOS
  - PowerShell installer for Windows
- Basic test suite with TDD approach
- GitHub Actions CI/CD pipeline
- MIT License

### Project Structure
- Clean architecture with separation of concerns
- Modular design with internal packages:
  - `api` - Cloudflare API client
  - `cmd` - CLI commands
  - `config` - Configuration management
  - `output` - Output formatting

### Documentation
- Comprehensive README with installation and usage instructions
- Code documentation and comments
- Example usage in help text

## Version History

- **v0.1.0** (2024-03-05) - Initial release with basic functionality
- **Unreleased** - Major improvements to developer experience and CLI usability

---

## Release Notes Format

Each release should include:

### Categories
- **Added** - New features
- **Changed** - Changes in existing functionality
- **Deprecated** - Soon-to-be removed features
- **Removed** - Removed features
- **Fixed** - Bug fixes
- **Security** - Vulnerability fixes

### Versioning
- **Major** (X.0.0) - Incompatible API changes
- **Minor** (0.X.0) - Added functionality in a backwards compatible manner
- **Patch** (0.0.X) - Backwards compatible bug fixes

[Unreleased]: https://github.com/PeterHiroshi/cfmon/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/PeterHiroshi/cfmon/compare/v0.3.1...v0.4.0
[v0.1.0]: https://github.com/PeterHiroshi/cfmon/releases/tag/v0.1.0