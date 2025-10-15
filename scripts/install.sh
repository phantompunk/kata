#!/bin/sh

set -e

BINARY_NAME="kata"
INSTALL_PATH="${INSTALL_PATH:-$HOME/.local/bin}"

RED="\033[31m"
BLUE="\033[34m"
GREEN="\033[32m"
YELLOW="\033[33m"
RESET="\033[0m"

info() {
    printf "${BLUE}==>${RESET} %s\n" "$1"
}

success() {
    printf "${GREEN}✓${RESET} %s\n" "$1"
}

error() {
    printf "${RED}𝘹${RESET} %s\n" "$1" >&2
}

warn() {
    printf "${YELLOW}!RESET} %s\n" "$1"
}

detect_os() {
    case $(uname -s) in    
        Linux*) echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *) error "Unsupported OS: $(uname -s)"; exit 1 ;;
    esac
}

detect_arch() {
    case $(uname -m) in    
        amd64|x86_64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

fetch() {
    local file="$1"
    local url="$2"

    if command_exists curl; then
        curl -fsSL -o "$file" "$url"
    elif command_exists wget; then
        wget -q -O "$file" "$url"
    else
        error "Can't find curl or wget, can't download package"
        exit 1
    fi
}

get_latest_version() {
    local url="https://api.github.com/repos/phantompunk/kata/releases/latest"
    local version

    if command_exists curl; then
        version=$(curl -fsSL "$url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command_exists wget; then
        version=$(wget -qO- "$url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    if [ -z "$version" ]; then
        error "Could not determine latest version from ${url}"
        exit 1
    fi

    echo "$version"
}

info "Installing ${BINARY_NAME}..."
OS=$(detect_os)
ARCH=$(detect_arch)
info "Detected: ${OS}/${ARCH}"

info "Fetching latest release..."
VERSION=$(get_latest_version)
info "Found version: ${VERSION}"

ARCHIVE="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/phantompunk/kata/releases/download/${VERSION}/${ARCHIVE}"

info "Downloading kata..."

temp_dir=$(mktemp -dt kata.XXXXXX)
trap 'rm -rf "$temp_dir"' EXIT INT TERM
cd "$temp_dir"

if ! fetch kata.tar.gz "$DOWNLOAD_URL"; then
    error "Could not download tarball from ${DOWNLOAD_URL}"
    exit 1
fi

if ! tar -xzf kata.tar.gz; then
    error "Failed to extract tarball"
    exit 1
fi

if [ ! -f "$BINARY_NAME" ]; then
    error "Binary ${BINARY_NAME} not found"
    exit 1
fi

info "Installing to ${INSTALL_PATH}..."
mkdir -p "$INSTALL_PATH"

if ! mv "$BINARY_NAME" "${INSTALL_PATH}/${BINARY_NAME}"; then
    error "Failed to move binary to ${INSTALL_PATH}"
    exit 1
fi

chmod +x "${INSTALL_PATH}/${BINARY_NAME}"
success "Installed ${BINARY_NAME} to ${INSTALL_PATH}/${BINARY_NAME}"

case ":$PATH:" in 
    *":${INSTALL_PATH}:"*)
        success "Installation complete!"
        info "Run: ${BINARY_NAME} --help"
        ;;
    *)
        success "Installation complete!"
        warn "${INSTALL_PATH} is not in your PATH"
        info "Add it by running:"
        echo "    echo 'export PATH=\"\$PATH:${INSTALL_PATH}\"' >> ~/.bashrc"
        echo "    source ~/.bashrc"
        ;;
esac
