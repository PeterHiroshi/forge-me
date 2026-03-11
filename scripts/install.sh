#!/bin/bash
set -e

# cfmon installer for macOS and Linux
# Usage: curl -sSL https://raw.githubusercontent.com/PeterHiroshi/cfmon/main/scripts/install.sh | bash
# Options:
#   VERSION=v0.2.0 - Install specific version
#   INSTALL_DIR=/custom/path - Custom installation directory

REPO="PeterHiroshi/cfmon"
BINARY_NAME="cfmon"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
RESET='\033[0m'
BOLD='\033[1m'

# Helper functions
print_banner() {
  echo -e "${CYAN}${BOLD}"
  echo "  ___  ___  __  __  ___  _  _  "
  echo " / __|| __||  \/  |/ _ \| \| | "
  echo "| (__ | _| | |\/| | (_) | .\  | "
  echo " \___||_|  |_|  |_|\___/|_|\_| "
  echo "                                "
  echo -e "${RESET}${BLUE}Cloudflare Workers/Containers CLI Monitor${RESET}"
  echo ""
}

print_error() {
  echo -e "${RED}✗ Error:${RESET} $1" >&2
}

print_warning() {
  echo -e "${YELLOW}⚠ Warning:${RESET} $1"
}

print_success() {
  echo -e "${GREEN}✓${RESET} $1"
}

print_info() {
  echo -e "${BLUE}ℹ${RESET} $1"
}

print_step() {
  echo -e "${MAGENTA}➜${RESET} $1..."
}

# Detect OS and architecture
detect_platform() {
  local os=$(uname -s | tr '[:upper:]' '[:lower:]')
  local arch=$(uname -m)

  case "$os" in
    darwin)
      OS="Darwin"
      OS_DISPLAY="macOS"
      ;;
    linux)
      OS="Linux"
      OS_DISPLAY="Linux"
      # Try to detect distribution
      if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS_DISPLAY="$NAME"
      fi
      ;;
    mingw*|msys*|cygwin*)
      print_error "Windows detected. Please use the PowerShell installer:"
      echo "  irm https://raw.githubusercontent.com/PeterHiroshi/cfmon/main/scripts/install.ps1 | iex"
      exit 1
      ;;
    *)
      print_error "Unsupported OS: $os"
      exit 1
      ;;
  esac

  case "$arch" in
    x86_64|amd64)
      ARCH="x86_64"
      ARCH_DISPLAY="x86_64"
      ;;
    aarch64|arm64)
      ARCH="arm64"
      ARCH_DISPLAY="ARM64"
      ;;
    i386|i686)
      ARCH="i386"
      ARCH_DISPLAY="x86"
      ;;
    armv6l|armv7l)
      ARCH="armv${arch:4:1}"
      ARCH_DISPLAY="ARM v${arch:4:1}"
      ;;
    *)
      print_error "Unsupported architecture: $arch"
      exit 1
      ;;
  esac
}

# Check dependencies
check_dependencies() {
  local missing=""

  for cmd in curl tar; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      missing="$missing $cmd"
    fi
  done

  if [ -n "$missing" ]; then
    print_error "Missing required dependencies:$missing"
    print_info "Please install them and try again"
    exit 1
  fi
}

# Get version (latest or specified)
get_version() {
  if [ -n "${VERSION:-}" ]; then
    # Version specified by user
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      print_warning "Invalid version format. Using latest version instead"
      VERSION=""
    fi
  fi

  if [ -z "${VERSION:-}" ]; then
    print_step "Fetching latest version"
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
      print_error "Failed to get latest version from GitHub"
      print_info "You can specify a version manually: VERSION=v0.1.0 $0"
      exit 1
    fi
  fi
}

# Determine install directory
determine_install_dir() {
  # Use custom dir if specified
  if [ -n "${INSTALL_DIR:-}" ]; then
    return
  fi

  # Check common install locations
  if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  elif [ -d "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
  elif [ -d "$HOME/bin" ]; then
    INSTALL_DIR="$HOME/bin"
  else
    INSTALL_DIR="/usr/local/bin"
  fi
}

# Download with progress
download_file() {
  local url="$1"
  local output="$2"
  local description="$3"

  print_step "Downloading $description"

  if command -v curl >/dev/null 2>&1; then
    if ! curl -fsSL --progress-bar "$url" -o "$output"; then
      return 1
    fi
  elif command -v wget >/dev/null 2>&1; then
    if ! wget -q --show-progress "$url" -O "$output"; then
      return 1
    fi
  else
    print_error "Neither curl nor wget found"
    return 1
  fi
  return 0
}

# Main installation
main() {
  print_banner

  # Check dependencies
  check_dependencies

  # Detect platform
  detect_platform
  print_info "Detected platform: ${BOLD}$OS_DISPLAY $ARCH_DISPLAY${RESET}"

  # Get version
  get_version
  print_info "Installing version: ${BOLD}$VERSION${RESET}"

  # Determine install directory
  determine_install_dir
  print_info "Install directory: ${BOLD}$INSTALL_DIR${RESET}"

  # Construct download URLs
  # GoReleaser archives use version without "v" prefix
  ARCHIVE_VERSION="${VERSION#v}"
  ARCHIVE="${BINARY_NAME}_${ARCHIVE_VERSION}_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
  CHECKSUM_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"

  # Create temp directory
  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT
  cd "$TMP_DIR"

  # Download archive
  if ! download_file "$DOWNLOAD_URL" "$ARCHIVE" "cfmon binary"; then
    print_error "Failed to download cfmon binary"
    print_info "URL: $DOWNLOAD_URL"
    exit 1
  fi

  # Download and verify checksum
  if download_file "$CHECKSUM_URL" "checksums.txt" "checksums"; then
    print_step "Verifying checksum"

    expected_checksum=$(grep "$ARCHIVE" checksums.txt | awk '{print $1}')

    if command -v sha256sum >/dev/null 2>&1; then
      actual_checksum=$(sha256sum "$ARCHIVE" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
      actual_checksum=$(shasum -a 256 "$ARCHIVE" | awk '{print $1}')
    else
      print_warning "Cannot verify checksum (no sha256sum or shasum found)"
      actual_checksum=""
    fi

    if [ -n "$actual_checksum" ]; then
      if [ "$expected_checksum" = "$actual_checksum" ]; then
        print_success "Checksum verified"
      else
        print_error "Checksum verification failed"
        print_info "Expected: $expected_checksum"
        print_info "Actual:   $actual_checksum"
        exit 1
      fi
    fi
  else
    print_warning "Could not download checksums file, skipping verification"
  fi

  # Extract archive
  print_step "Extracting archive"
  if ! tar -xzf "$ARCHIVE"; then
    print_error "Failed to extract archive"
    exit 1
  fi

  # Install binary
  print_step "Installing cfmon to $INSTALL_DIR"

  # Create install directory if it doesn't exist
  if [ ! -d "$INSTALL_DIR" ]; then
    if ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
      print_info "Creating $INSTALL_DIR requires sudo"
      sudo mkdir -p "$INSTALL_DIR"
    fi
  fi

  # Move binary to install directory
  if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
  else
    print_info "Installing to $INSTALL_DIR requires sudo"
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
  fi

  # Make sure it's executable
  if [ -w "$INSTALL_DIR/$BINARY_NAME" ]; then
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
  else
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
  fi

  # Install shell completions if directory exists
  if [ -d "/usr/local/share/bash-completion/completions" ] && [ -f "completions/cfmon.bash" ]; then
    print_step "Installing bash completions"
    sudo cp completions/cfmon.bash /usr/local/share/bash-completion/completions/cfmon 2>/dev/null || true
  fi

  if [ -d "/usr/local/share/zsh/site-functions" ] && [ -f "completions/cfmon.zsh" ]; then
    print_step "Installing zsh completions"
    sudo cp completions/cfmon.zsh /usr/local/share/zsh/site-functions/_cfmon 2>/dev/null || true
  fi

  # Verify installation
  echo ""
  if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    installed_version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
    print_success "${GREEN}${BOLD}cfmon installed successfully!${RESET}"
    echo ""
    echo -e "${CYAN}${BOLD}Quick Start:${RESET}"
    echo -e "  ${BOLD}1.${RESET} Set your Cloudflare API token:"
    echo -e "     ${CYAN}cfmon login <your-token>${RESET}"
    echo -e "  ${BOLD}2.${RESET} Check system status:"
    echo -e "     ${CYAN}cfmon doctor${RESET}"
    echo -e "  ${BOLD}3.${RESET} List your resources:"
    echo -e "     ${CYAN}cfmon containers list${RESET}"
    echo -e "     ${CYAN}cfmon workers list${RESET}"
    echo -e "  ${BOLD}4.${RESET} Get help:"
    echo -e "     ${CYAN}cfmon help${RESET}"
  else
    print_warning "Installation complete, but $BINARY_NAME is not in PATH"
    print_info "Add $INSTALL_DIR to your PATH:"
    echo ""
    echo "  For bash:  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
    echo "  For zsh:   echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc"
    echo "  For fish:  fish_add_path $INSTALL_DIR"
    echo ""
    print_info "Then reload your shell or run: source ~/.bashrc (or ~/.zshrc)"
  fi

  echo ""
  echo -e "${BLUE}${BOLD}Additional Installation Methods:${RESET}"
  echo -e "  ${BOLD}Homebrew${RESET} (macOS/Linux):"
  echo -e "    ${CYAN}brew tap PeterHiroshi/cfmon${RESET}"
  echo -e "    ${CYAN}brew install cfmon${RESET}"
  echo ""
  echo -e "  ${BOLD}From source${RESET}:"
  echo -e "    ${CYAN}git clone https://github.com/PeterHiroshi/cfmon${RESET}"
  echo -e "    ${CYAN}cd cfmon && make install${RESET}"
  echo ""
  echo -e "${GREEN}${BOLD}Thank you for installing cfmon!${RESET}"
  echo -e "Documentation: ${BLUE}https://github.com/PeterHiroshi/cfmon${RESET}"
}

# Run main function
main "$@"
