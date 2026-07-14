#!/usr/bin/env bash

# No set -e — installer should be resilient to failures

API_KEY="${TRASHNEURONS_API_KEY:-}"
REPO_URL="https://github.com/velez1337fn/trashtalk-neurons"
RELEASE_URL="https://api.github.com/repos/velez1337fn/trashtalk-neurons/releases/latest"
INSTALL_DIR="$HOME/.trashneurons"
BIN_NAME="trashneurons-cli"
GITHUB_TOKEN="${GITHUB_TOKEN:-}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Detect if running in a terminal
if [ -t 0 ]; then
    IS_TERMINAL=true
else
    IS_TERMINAL=false
fi

print_logo() {
    if [ "$IS_TERMINAL" = true ]; then
        echo -e "${CYAN}"
        cat << 'LOGO'
   .___                 __         .__  .__                              ____    _______      _______
   |   | ____   _______/  |______  |  | |  |   ___________      ___  __ /_   |   \   _  \     \   _  \
   |   |/    \ /  ___/\   __\__  \ |  | |  | _/ __ \_  __ \     \  \/ /  |   |   /  /_\  \    /  /_\  \
   |   |   |  \\___ \  |  |  / __ \|  |_|  |_\  ___/|  | \/      \   /   |   |   \  \_/   \   \  \_/   \
   |___|___|  /____  > |__| (____  /____/____/\___  >__|          \_/    |___| /\ \_____  / /\ \_____  /
            \/     \/            \/               \/                           \/       \/  \/       \/
LOGO
        echo -e "${NC}"
    fi
}

detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) echo "amd64" ;;
    esac
}

install_cli() {
    local arch
    arch=$(detect_arch)
    local bin_file="trashneurons-cli-${arch}"
    local tmp_bin="/tmp/${bin_file}"

    echo "Detecting system... linux-${arch}"
    echo "Installing to ${INSTALL_DIR}/bin/"

    mkdir -p "$INSTALL_DIR/bin"

    echo "Fetching latest release..."

    local latest_release
    if [ -n "$GITHUB_TOKEN" ]; then
        latest_release=$(curl -s -H "Authorization: token $GITHUB_TOKEN" "$RELEASE_URL")
    else
        latest_release=$(curl -s "$RELEASE_URL")
    fi

    local asset_url
    asset_url=$(echo "$latest_release" | grep -o '"browser_download_url": "[^"]*'"${arch}"'[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -z "$asset_url" ]; then
        echo "ERROR: No release found for arch ${arch}"
        echo "Run 'git tag v1.x.x && git push origin v1.x.x' to trigger a build."
        exit 1
    fi

    echo "Downloading from: $asset_url"
    curl -L -o "$tmp_bin" "$asset_url" 2>/dev/null || {
        echo "ERROR: Download failed"
        exit 1
    }

    chmod +x "$tmp_bin"

    cp "$tmp_bin" "$INSTALL_DIR/bin/${BIN_NAME}"
    rm -f "$tmp_bin"

    if [ ! -L "/usr/local/bin/${BIN_NAME}" ] && [ ! -f "/usr/local/bin/${BIN_NAME}" ]; then
        sudo ln -sf "$INSTALL_DIR/bin/${BIN_NAME}" "/usr/local/bin/${BIN_NAME}" 2>/dev/null || true
    fi

    echo ""
    echo "Installation complete!"
    echo "Files located in: ${INSTALL_DIR}"
    echo "Run: trashneurons-cli"
}

show_welcome_screen() {
    clear
    print_logo

    local user
    user=$(whoami)
    local arch
    arch=$(detect_arch)
    local sys_info="linux-${arch}"

    printf "${GREEN}┌─────────────────────┐${NC}%-56s${GREEN}┌──────────────────────────────────────────┐${NC}\n" ""
    printf "${GREEN}│${NC} ${CYAN}welcome, ${user}!${NC}       ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} $(printf ' %.0s' {1..58})${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}${GREEN}├─────────────────────┤${NC}$(printf ' %.0s' {1..56})${GREEN}├──────────────────────────────────────────┤${NC}\n"
    printf "${GREEN}│${NC} ${BLUE}matrix in this box${NC} ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} $(printf ' %.0s' {1..58})${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}${GREEN}├─────────────────────┐${NC}%-56s${GREEN}├──────────────────────────────────────────┤${NC}\n" ""
    printf "${GREEN}│${NC} ${YELLOW}detected sys:${NC} ${sys_info}${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} $(printf ' %.0s' {1..58})${GREEN}│${NC}\n"
    printf "${GREEN}│${NC} ${YELLOW}installer status:${NC} ready${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} $(printf ' %.0s' {1..58})${GREEN}│${NC}\n"
    printf "${GREEN}│${NC} ${YELLOW}waiting for input...${NC}${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}are ya winning son?     ***   ***${NC}     ${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}                     ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}    ─┬────┐${NC}             ***** *****       ${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}                     ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}     └─┬──┘${NC}              *********        ${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}                     ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}   ┌───┼──┐${NC}               *******         ${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}                     ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}   │   │  │${NC}                *****          ${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}                     ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}     ┌─┼─┐${NC}                  ***           ${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}                     ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}     │   │${NC}                   *            ${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}                     ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}     └─┬─┬┘${NC}                    ──           ${GREEN}│${NC}\n"
    printf "${GREEN}│${NC}                     ${GREEN}│${NC}$(printf ' %.0s' {1..56})${GREEN}│${NC} ${RED}       └─┘${NC}                         ***        ${GREEN}│${NC}\n"
    printf "${GREEN}└─────────────────────┴─────────────────────────────────────────────────────────┴─────┴───┴────────────────────────────────┘${NC}\n"
    echo ""
    echo -e "  ${GREEN}1.${NC} Install TrashNeurons CLI"
    echo -e "  ${GREEN}2.${NC} Exit"
    echo ""
    echo -ne "  ${YELLOW}Choose option [1-2]: ${NC}"
    read -r choice < /dev/tty 2>/dev/null || read -r choice

    case "${choice}" in
        1) install_cli ;;
        2) exit 0 ;;
        *) echo -e "${RED}Invalid option. Exiting.${NC}"; exit 1 ;;
    esac
}

show_welcome_screen
