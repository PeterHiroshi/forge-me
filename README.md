<div align="center">

# 🔨 cfmon

**A powerful CLI for Cloudflare Workers and Containers monitoring**

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/PeterHiroshi/cfmon)](https://github.com/PeterHiroshi/cfmon/releases)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen.svg)](https://github.com/PeterHiroshi/cfmon/actions)
[![Coverage](https://img.shields.io/badge/coverage->90%25-brightgreen.svg)](https://github.com/PeterHiroshi/cfmon)

</div>

## 📖 Overview

**cfmon** is a fast, feature-rich CLI tool for monitoring and managing your Cloudflare resources. Built for developers and DevOps engineers who prefer powerful command-line tools over web dashboards, it provides instant access to your Workers and Containers with detailed metrics, advanced filtering, and automation support.

## 📑 Table of Contents

- [Overview](#-overview)
- [Quick Start](#-quick-start)
- [What's New](#-whats-new)
- [Features](#-features)
- [Installation](#-installation)
- [Usage](#-usage)
- [Dashboard Guide](#-dashboard-guide)
- [Command Reference](#-command-reference)
- [Examples](#-examples)
- [Configuration](#️-configuration)
- [Development](#️-development)
- [Contributing](#-contributing)
- [Support](#-support)

---

## ⚡ Quick Start

```bash
# Install cfmon
curl -sSL https://raw.githubusercontent.com/PeterHiroshi/cfmon/main/scripts/install.sh | bash

# Save your Cloudflare API token
cfmon login YOUR_API_TOKEN

# Check system health
cfmon doctor

# Launch interactive dashboard
cfmon dashboard YOUR_ACCOUNT_ID

# List containers with filtering and sorting
cfmon containers list YOUR_ACCOUNT_ID --filter "prod" --sort cpu --limit 10

# One-shot health check with threshold alerts
cfmon check YOUR_ACCOUNT_ID --cpu-threshold 70

# Stream real-time logs from a worker
cfmon tail my-worker --format pretty

# View configuration
cfmon config show
```

---

## 🎉 What's New

### Version 0.4.0

- **📡 Real-time Log Streaming (`cfmon tail`)** — Stream live logs from Cloudflare Workers and Containers
  - WebSocket-based real-time log streaming via Cloudflare Tail API
  - 12+ filtering flags: `--status`, `--method`, `--search`, `--ip`, `--header`, `--since`, `--sample-rate`, `--max-events`
  - Multiple output formats: pretty (colored), JSON, compact
  - Automatic WebSocket reconnection on disconnect
  - Clean session management with graceful shutdown
  - Client-side and server-side filtering
  - More powerful than `wrangler tail`

### Version 0.3.0

- **📺 Interactive TUI Dashboard** — Real-time terminal UI with 4 tabs (Overview, Workers, Containers, Alerts)
  - Health gauge with color-coded scoring
  - Workers/Containers tables with live metrics, ASCII resource bars
  - Alerts tab with severity-based threshold monitoring and event log
  - Keyboard navigation (j/k, Tab, 1-4), search/filter (`/`), detail view (Enter)
  - Mouse scroll support, help overlay (`?`)
  - Auto-refresh with configurable interval (`--refresh`)

### Version 0.2.0

- **🩺 Doctor Command**: Comprehensive system health checks
- **🎯 Advanced Filtering**: Filter resources by name with `--filter`
- **📊 Sorting**: Sort by CPU, memory, requests with `--sort`
- **🔢 Limiting**: Control output size with `--limit`
- **⚙️ Config Management**: New `config show` and `config path` commands
- **🐛 Debug Mode**: Verbose output with `-v` flag
- **⏱️ Custom Timeouts**: Set API timeout with `--timeout`
- **📝 Rich Help**: Beautiful, example-filled help with `cfmon help`
- **🔨 Makefile**: Easy development with `make build`, `make test`, etc.
- **📦 Cross-Platform Releases**: Automated releases with GoReleaser
- **🧪 Comprehensive Testing**: >90% test coverage with integration tests

---

## ✨ Features

### Core Features

#### 📡 **Real-time Log Streaming** ⭐ NEW
- **`cfmon tail`** — Stream live logs from Workers and Containers
- **12+ flags** for precise filtering (status, method, IP, headers, search text)
- **3 output formats**: pretty (colored), JSON, compact
- **WebSocket auto-reconnect** on connection drop
- **Client-side filtering**: `--search`, `--since`, `--max-events`
- **Server-side filtering**: `--status`, `--method`, `--ip`, `--header`, `--sample-rate`
- More powerful than `wrangler tail`

#### 📺 **Interactive TUI Dashboard** ⭐ NEW
- **Real-time monitoring** with auto-refresh (configurable interval)
- **4 tabs**: Overview, Workers, Containers, Alerts
- **Overview tab**: Health gauge (0–100), resource summaries, alert counts
- **Workers tab**: Scrollable table with CPU, requests, errors, success rate
- **Containers tab**: Scrollable table with CPU/memory ASCII bars, status indicators
- **Alerts tab**: Threshold-based alerts with severity levels, event timeline
- **Keyboard-driven**: j/k navigation, Tab switching, `/` filter, Enter for detail, `?` for help
- **Mouse support**: Scroll wheel for table navigation
- **Detail view**: Press Enter on any row for expanded resource information
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)

#### 📦 **Container Management**
- **List** all containers with resource metrics
- **Status** command for detailed container information
- **Filter** containers by name pattern
- **Sort** by CPU, memory, or request count
- **Limit** output to top N results

#### ⚡ **Worker Monitoring**
- **List** all workers with performance metrics
- **Status** command for individual worker details
- **Filter** workers by name substring
- **Sort** by various metrics (CPU, requests, errors)
- **Track** success rates and error counts

#### 🩺 **System Health Checks**
- **Doctor** command for comprehensive diagnostics
- **Health** command for 0-100 point scoring
- **Check** command for threshold-based alerts with exit codes
- Verify API token validity
- Check network connectivity
- Test Cloudflare API access
- Validate configuration

#### ⚙️ **Configuration Management**
- **Show** current configuration with masked secrets
- **Path** command to locate config file
- Support for environment variables
- Flexible token management

### Output & Formatting

#### 🎨 **Multiple Output Formats**
- **Table**: Beautiful colored tables (default)
- **JSON**: Machine-readable for automation
- **No-color**: Plain text for logs/CI

#### 🔍 **Advanced Filtering & Sorting**
```bash
# Filter by name pattern
cfmon containers list <account> --filter "prod"

# Sort by CPU usage (descending)
cfmon workers list <account> --sort cpu

# Combine filter, sort, and limit
cfmon containers list <account> --filter "api" --sort memory --limit 5
```

### Developer Experience

#### 🐚 **Shell Completions**
- Bash, Zsh, Fish, PowerShell support
- Tab completion for commands and flags
- Install with: `cfmon completion <shell>`

#### 🐛 **Debug & Verbose Mode**
- Use `-v` or `--verbose` for debug output
- Detailed error messages with suggestions
- API request/response logging

#### ⏱️ **Timeout Control**
- Customize API timeout: `--timeout 30s`
- Prevent hanging on slow connections
- Configurable per command

---

## 🚀 Installation

### Quick Install (Recommended)

```bash
# Linux/macOS
curl -sSL https://raw.githubusercontent.com/PeterHiroshi/cfmon/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/PeterHiroshi/cfmon/main/scripts/install.ps1 | iex
```

### Package Managers

#### Homebrew (macOS/Linux)
```bash
brew tap PeterHiroshi/cfmon
brew install cfmon
```

#### From Source
```bash
git clone https://github.com/PeterHiroshi/cfmon
cd cfmon
make install
```

#### Download Binary
Download pre-built binaries from [Releases](https://github.com/PeterHiroshi/cfmon/releases)

---

## 📘 Usage

### Initial Setup

1. **Get your Cloudflare API Token**
   - Go to [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens)
   - Create token with `Account:Read` and `Workers Scripts:Read` permissions

2. **Configure cfmon**
   ```bash
   # Save your token
   cfmon login <your-api-token>

   # Verify setup
   cfmon doctor
   ```

3. **Find your Account ID**
   - Visit [Cloudflare Dashboard](https://dash.cloudflare.com)
   - Copy Account ID from the right sidebar

### Basic Commands

```bash
# List containers
cfmon containers list <account-id>

# List workers
cfmon workers list <account-id>

# Get container status
cfmon containers status <account-id> <container-id>

# Get worker status
cfmon workers status <account-id> <worker-name>
```

### Advanced Usage

```bash
# Filter and sort containers
cfmon containers list <account-id> --filter "prod" --sort cpu --limit 10

# JSON output for automation
cfmon workers list <account-id> --format json | jq '.[] | select(.cpu_ms > 1000)'

# Verbose mode for debugging
cfmon containers list <account-id> -v

# Custom timeout
cfmon workers list <account-id> --timeout 60s

# Use different token
cfmon containers list <account-id> --token <other-token>
```

---

## 📺 Dashboard Guide

The interactive TUI dashboard provides a unified real-time view of your entire Cloudflare account. It is the primary way to monitor resources at a glance.

### Launching the Dashboard

```bash
# With explicit account ID
cfmon dashboard <account-id>

# Using default account (set via cfmon accounts set-default)
cfmon dashboard

# Custom refresh interval (default 30s, minimum 5s)
cfmon dashboard <account-id> --refresh 10s
```

### Dashboard Tabs

#### Tab 1 — Overview

The landing tab shows a high-level health summary:

| Element | Description |
|---------|-------------|
| **Health Gauge** | ASCII progress bar with 0–100 score. Green ≥75, yellow ≥50, red <50 |
| **Workers Summary** | Total workers count, aggregate CPU/request/error metrics |
| **Containers Summary** | Total containers count, aggregate CPU/memory usage |
| **Alert Count** | Number of active warnings and critical alerts |

#### Tab 2 — Workers

Scrollable table of all Workers with columns:

| Column | Description |
|--------|-------------|
| Name | Worker script name |
| Status | Running status with color indicator |
| CPU (ms) | CPU time in milliseconds |
| Requests | Total request count |
| Errors | Error count |
| Success Rate | Percentage with color coding (green ≥99%, yellow ≥95%, red <95%) |

#### Tab 3 — Containers

Scrollable table of all Containers with columns:

| Column | Description |
|--------|-------------|
| Name / ID | Container identifier |
| Status | Running status with color indicator |
| CPU | Usage with ASCII bar visualization |
| Memory | Usage with ASCII bar visualization |
| Requests | Total request count |

#### Tab 4 — Alerts

Threshold-based alert monitoring:

- Alerts are evaluated against configurable thresholds (CPU, memory, error rate)
- Severity levels: **OK** (green), **Warning** (yellow), **Critical** (red)
- **Event log**: Timeline of alert state changes (new alerts, resolved alerts)
- Filter alerts by resource name with `/`
- Tab badge shows alert count when alerts are active

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` / `Shift-Tab` | Switch between tabs |
| `1` `2` `3` `4` | Jump directly to a tab |
| `j` / `↓` | Move cursor down / scroll down |
| `k` / `↑` | Move cursor up / scroll up |
| `Enter` | Open detail view for selected row |
| `Esc` | Close detail / dismiss filter / exit |
| `/` | Enter filter mode (Workers, Containers, Alerts tabs) |
| `r` | Force refresh data |
| `?` | Toggle help overlay |
| `q` | Quit dashboard |
| Mouse wheel | Scroll tables up/down |

### Filter Mode

Press `/` on any list tab to activate filter mode:

1. Type your search query — the table filters in real-time
2. Press `Enter` to confirm the filter and return to normal navigation
3. Press `Esc` to cancel and clear the filter
4. The status bar shows the active filter when set

Filter matches resource names as case-insensitive substrings.

### Detail View

Press `Enter` on a selected row in Workers or Containers tab:

- Shows expanded information for the selected resource
- Displays all available metrics and metadata
- Press `Esc` to return to the table view

---

## 📚 Command Reference

### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Output format (table, json) | table |
| `--verbose` | `-v` | Enable debug output | false |
| `--timeout` | | API request timeout | 30s |
| `--no-color` | | Disable colored output | false |
| `--config` | | Config file path | ~/.cfmon/config.yaml |
| `--token` | | Override API token | |
| `--help` | `-h` | Show help | |
| `--version` | | Show version | |

### Commands

#### `cfmon dashboard [account-id]`
Launch interactive TUI dashboard

**Flags:**
- `--refresh <duration>`: Auto-refresh interval (default: 30s, min: 5s)

**Controls:** Tab/1-4 to switch tabs, j/k to navigate, `/` to filter, Enter for detail, `?` for help, `q` to quit.

#### `cfmon help [command]`
Display detailed help with examples

#### `cfmon doctor`
Run system diagnostics
- Check Go runtime
- Verify configuration
- Test API token
- Check network connectivity

#### `cfmon login <token>`
Save Cloudflare API token

#### `cfmon config show`
Display current configuration (tokens masked)

#### `cfmon config path`
Show configuration file location

#### `cfmon containers list <account-id>`
List all containers

**Flags:**
- `--filter <pattern>`: Filter by name substring
- `--sort <field>`: Sort by name, cpu, memory, requests
- `--limit <n>`: Limit results

#### `cfmon containers status <account-id> <container-id>`
Get detailed container information

#### `cfmon workers list <account-id>`
List all workers

**Flags:**
- `--filter <pattern>`: Filter by name substring
- `--sort <field>`: Sort by name, cpu, requests, errors
- `--limit <n>`: Limit results

#### `cfmon workers status <account-id> <worker-name>`
Get detailed worker information

#### `cfmon check [account-id]`
One-shot health check with threshold-based alerts

**Flags:**
- `--cpu-threshold <percent>`: CPU usage warning threshold (default: 80)
- `--memory-threshold <percent>`: Memory usage warning threshold (default: 85)
- `--error-threshold <percent>`: Error rate warning threshold (default: 2)

**Exit codes:** 0 = healthy, 1 = warnings, 2 = critical

#### `cfmon tail [account-id] <worker-name>`
Stream real-time logs from a Cloudflare Worker or Container

**Flags:**
- `--format, -f <format>`: Output format: pretty, json, compact (default: pretty)
- `--status <codes>`: Filter by HTTP status: ok, error, or status codes
- `--method <methods>`: Filter by HTTP method: GET, POST, etc
- `--search <text>`: Filter logs containing this string
- `--ip <addresses>`: Filter by client IP address
- `--header <key:value>`: Filter by request header
- `--sample-rate <rate>`: Sampling rate 0.0-1.0 (default: 1.0)
- `--max-events, -n <count>`: Stop after N events
- `--since <duration>`: Only show events after duration (e.g. 5m, 1h)
- `--no-color`: Disable colored output
- `--include-logs`: Show console.log() output (default: true)
- `--include-exceptions`: Show exceptions (default: true)

**Examples:**
```bash
# Stream all logs with colored output
cfmon tail my-worker

# JSON format for piping
cfmon tail my-worker --format json | jq .

# Filter errors only
cfmon tail my-worker --status error

# Search for specific text, limit to 100 events
cfmon tail my-worker --search "timeout" --max-events 100

# Filter by method and IP
cfmon tail my-worker --method POST --ip 1.2.3.4
```

#### `cfmon completion <shell>`
Generate shell completion script

**Supported shells:**
- bash
- zsh
- fish
- powershell

---

## 💡 Examples

### Launch the Dashboard

```bash
# Quick start — interactive monitoring
cfmon dashboard <account-id>

# Faster updates for incident response
cfmon dashboard <account-id> --refresh 5s

# Using default account
cfmon accounts set-default <account-id>
cfmon dashboard
```

### Monitor High CPU Usage

```bash
# Find top 5 CPU-consuming containers
cfmon containers list <account> --sort cpu --limit 5

# Monitor specific worker
watch -n 5 'cfmon workers status <account> api-gateway'
```

### Automation & Scripting

```bash
#!/bin/bash
# Alert when worker CPU exceeds threshold

THRESHOLD=5000
WORKERS=$(cfmon workers list <account> --format json)

echo "$WORKERS" | jq -r '.[] | select(.cpu_ms > '$THRESHOLD') |
  "Alert: \(.name) is using \(.cpu_ms)ms CPU"'
```

### Multi-Account Management

```bash
# Use environment variables for different accounts
export PROD_TOKEN="prod-token-xxx"
export DEV_TOKEN="dev-token-yyy"

# Check production
cfmon containers list prod-account --token $PROD_TOKEN

# Check development
cfmon containers list dev-account --token $DEV_TOKEN
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Check Cloudflare Workers
  run: |
    cfmon doctor
    cfmon workers list ${{ secrets.CF_ACCOUNT_ID }} --format json > workers.json

    # Fail if any worker has errors
    if jq -e '.[] | select(.errors > 0)' workers.json; then
      echo "Workers with errors detected!"
      exit 1
    fi
```

---

## ⚙️ Configuration

### Configuration File

Location: `~/.cfmon/config.yaml` (or `$CFMON_CONFIG`)

```yaml
token: your-api-token-here
api_endpoint: https://api.cloudflare.com/client/v4  # optional
default_format: table  # or json
```

### Environment Variables

| Variable | Description | Priority |
|----------|-------------|----------|
| `CFMON_TOKEN` | API token | Highest |
| `CFMON_CONFIG` | Config file path | |
| `CFMON_FORMAT` | Default output format | |
| `CFMON_NO_COLOR` | Disable colors | |

Priority: Environment > Command Flag > Config File

---

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Make
- Git

### Setup

```bash
# Clone repository
git clone https://github.com/PeterHiroshi/cfmon
cd cfmon

# Install dependencies
make deps

# Run tests
make test

# Build binary
make build
```

### Makefile Targets

```bash
make help        # Show all targets
make build       # Build binary
make test        # Run unit tests
make coverage    # Generate coverage report
make lint        # Run linters
make install     # Install to GOPATH/bin
make clean       # Clean build artifacts
make release     # Build release binaries
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make integration-test

# Coverage report
make coverage
open coverage.html
```

### Project Structure

```
cfmon/
├── cmd/              # CLI commands
│   ├── root.go       # Root command
│   ├── dashboard.go  # Interactive TUI dashboard
│   ├── containers.go # Container commands
│   ├── workers.go    # Worker commands
│   ├── check.go      # Threshold-based health check
│   ├── doctor.go     # Doctor command
│   ├── config.go     # Config commands
│   └── help.go       # Help command
├── internal/         # Internal packages
│   ├── api/          # Cloudflare API client
│   ├── config/       # Configuration management
│   ├── dashboard/    # TUI dashboard (Bubble Tea)
│   │   ├── model.go      # Main model + Update loop
│   │   ├── types.go      # Tab IDs, data types, events
│   │   ├── fetcher.go    # Async data fetching
│   │   ├── workers.go    # Workers tab rendering
│   │   ├── containers.go # Containers tab rendering
│   │   ├── alerts.go     # Alerts tab + event log
│   │   ├── gauge.go      # ASCII health gauge
│   │   ├── filter.go     # Filter logic
│   │   ├── styles.go     # Lip Gloss styles
│   │   └── help.go       # Help overlay
│   ├── monitor/      # Threshold-based alert evaluation
│   └── output/       # Output formatting
├── skill/            # OpenClaw skill files
│   └── SKILL.md      # AI assistant skill definition
├── test/             # Test files
│   └── integration/  # Integration tests
├── scripts/          # Installation scripts
├── Makefile          # Build automation
├── .goreleaser.yml   # Release configuration
└── go.mod            # Go modules
```

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Contribution Guide

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests
5. Run `make check` to ensure quality
6. Commit (`git commit -m 'feat: add amazing feature'`)
7. Push (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Workflow

```bash
# Create branch
git checkout -b feature/new-command

# Make changes and test
make test

# Check code quality
make lint

# Build and test manually
make build
./cfmon doctor

# Commit and push
git add .
git commit -m "feat: add new command"
git push origin feature/new-command
```

---

## 📦 Releasing

Releases are automated with GoReleaser:

```bash
# Create a tag
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0

# GoReleaser will automatically:
# - Build binaries for all platforms
# - Create GitHub release
# - Update Homebrew formula
# - Generate changelog
```

---

## 🆘 Support

### Getting Help

- **Documentation**: Read this README and run `cfmon help`
- **Issues**: [GitHub Issues](https://github.com/PeterHiroshi/cfmon/issues)
- **Discussions**: [GitHub Discussions](https://github.com/PeterHiroshi/cfmon/discussions)

### Troubleshooting

#### Token Issues
```bash
# Verify token is set
cfmon config show

# Test token validity
cfmon doctor

# Re-login if needed
cfmon login <new-token>
```

#### Network Issues
```bash
# Check connectivity
cfmon doctor

# Increase timeout
cfmon workers list <account> --timeout 60s

# Use verbose mode for debugging
cfmon containers list <account> -v
```

#### Configuration Issues
```bash
# Show config location
cfmon config path

# Reset configuration
rm ~/.cfmon/config.yaml
cfmon login <token>
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [Cloudflare API v4](https://developers.cloudflare.com/api/)
- Inspired by the need for better CLI tools in the Cloudflare ecosystem

---

<div align="center">

**Made with ❤️ by developers, for developers**

[Report Bug](https://github.com/PeterHiroshi/cfmon/issues) • [Request Feature](https://github.com/PeterHiroshi/cfmon/issues) • [Star on GitHub](https://github.com/PeterHiroshi/cfmon)

</div>