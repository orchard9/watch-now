#!/bin/bash
# watch-now installer script
# Usage: curl -sSfL https://raw.githubusercontent.com/orchard9/watch-now/main/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
RESET='\033[0m'

# Helper functions
info() { echo -e "${BLUE}[INFO]${RESET} $1"; }
success() { echo -e "${GREEN}[✓]${RESET} $1"; }
error() { echo -e "${RED}[✗]${RESET} $1" >&2; }
warning() { echo -e "${YELLOW}[!]${RESET} $1"; }

# Main installation
main() {
    echo -e "${BOLD}Installing watch-now${RESET}"
    echo "===================="
    echo ""

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go 1.19 or later from https://golang.org/dl/"
        exit 1
    fi

    # Check Go version
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    info "Found Go version: ${go_version}"

    # Install using go install
    info "Installing watch-now..."
    
    if go install github.com/orchard9/watch-now@latest; then
        success "Successfully installed watch-now"
    else
        error "Failed to install watch-now"
        exit 1
    fi

    # Verify installation
    if command -v watch-now &> /dev/null; then
        success "watch-now is available in your PATH"
        echo ""
        watch-now --version
    else
        warning "watch-now was installed but not found in PATH"
        echo ""
        echo "Make sure $(go env GOPATH)/bin is in your PATH:"
        echo ""
        echo "  export PATH=\"\$PATH:$(go env GOPATH)/bin\""
        echo ""
        
        # Detect shell and suggest specific file
        local shell_config=""
        case "$SHELL" in
            */bash)
                shell_config="~/.bashrc or ~/.bash_profile"
                ;;
            */zsh)
                shell_config="~/.zshrc"
                ;;
            */fish)
                shell_config="~/.config/fish/config.fish"
                echo "For fish shell, use:"
                echo "  set -gx PATH \$PATH $(go env GOPATH)/bin"
                echo ""
                ;;
            *)
                shell_config="your shell configuration file"
                ;;
        esac
        
        if [[ "$SHELL" != */fish ]]; then
            echo "Add it to ${shell_config}:"
            echo "  echo 'export PATH=\"\$PATH:$(go env GOPATH)/bin\"' >> ${shell_config}"
        fi
    fi

    echo ""
    echo "To get started:"
    echo "  watch-now --init         # Generate configuration for your project"
    echo "  watch-now                # Start monitoring"
    echo "  watch-now --help         # Show help and configuration format"
    echo "  watch-now --show-examples # Show example configurations"
    echo ""
    success "Installation complete!"
}

# Run main function
main "$@"