#!/usr/bin/env bash
set -euo pipefail

APP=trashneurons-cli
MUTED='\033[0;2m'
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
ORANGE='\033[38;5;214m'
NC='\033[0m'

if [[ -t 1 ]]; then
    Bold_White='\033[1m'
    Bold_Green='\033[1;32m'
else
    Bold_White=''
    Bold_Green=''
fi

info() {
    echo -e "${Muted}$@ ${NC}"
}

info_bold() {
    echo -e "${Bold_White}$@ ${NC}"
}

success() {
    echo -e "${Green}$@ ${NC}"
}

error() {
    echo -e "${Red}$@ ${NC}" >&2
    exit 1
}

install_dir="$HOME/.trashneurons/bin"
mkdir -p "$install_dir"

raw_os=$(uname -s)
os=$(echo "$raw_os" | tr '[:upper:]' '[:lower:]')
case "$raw_os" in
  Darwin*) os="darwin" ;;
  Linux*) os="linux" ;;
  MINGW*|MSYS*|CYGWIN*) os="windows" ;;
esac

arch=$(uname -m)
case "$arch" in
  aarch64) arch="arm64" ;;
  x86_64) arch="amd64" ;;
esac

if [[ "$os" != "linux" && "$os" != "darwin" ]]; then
    error "Unsupported platform: $os/$arch"
fi

if [[ "$os" == "linux" ]]; then
    if ! command -v tar >/dev/null 2>&1; then
        error "'tar' is required but not installed"
    fi
fi

filename="${APP}-${os}-${arch}.tar.gz"
url="https://github.com/velez1337fn/trashtalk-neurons/releases/download/v1.0.0/$filename"
exe="$install_dir/$APP"

echo -e "${Blue}┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC} ${Green}>${NC} allocating path where app/cli been installed...                                                                    ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC}                 ${Green}found! path is:${NC} ${install_dir}                                                                            ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC} ${Green}>${NC} starting download via curl...                                                                                       ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"

tmp_dir="${TMPDIR:-/tmp}/trashneurons_install_$$"
mkdir -p "$tmp_dir"

curl --fail --location --progress-bar --output "$tmp_dir/$filename" "$url" ||
    error "Failed to download from $url"

tar -xzf "$tmp_dir/$filename" -C "$tmp_dir"

mv "$tmp_dir/$APP" "$exe"
chmod +x "$exe"
rm -rf "$tmp_dir"

echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC}    ${Green}installing your app/cli...${NC}                                                                                          ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC}  ${Green}finished install!${NC} all files located in ${install_dir}                                                                   ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}│${NC}  ${MUTED}for use:${NC} type \"trashneurons-cli\"                                                                                    ${Blue}│${NC}"
echo -e "${Blue}│${NC}                                                                                                                        ${Blue}│${NC}"
echo -e "${Blue}├──────────────────────┐${NC}                                                                                 ${Blue}├────────────────────────────────────────────────────────┤${NC}"
echo -e "${Blue}│${NC} ${RED}PRESS CTRL+C TO CLOSE INSTALLER${NC}   ${Blue}│${NC}                                                                                 ${Blue}│${NC}"
echo -e "${Blue}└──────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘${NC}"

current_shell=$(basename "$SHELL" 2>/dev/null || echo "bash")
case $current_shell in
    fish) config_file="$HOME/.config/fish/config.fish" ;;
    zsh) config_file="${ZDOTDIR:-$HOME}/.zshrc" ;;
    bash) config_file="$HOME/.bashrc" ;;
    *) config_file="$HOME/.bashrc" ;;
esac

if [[ -f "$config_file" ]] && [[ ":$PATH:" != *":$install_dir:"* ]]; then
    case $current_shell in
        fish) echo "fish_add_path $install_dir" >> "$config_file" ;;
        *) echo ""; echo "# trashneurons" >> "$config_file"; echo "export PATH=$install_dir:\$PATH" >> "$config_file" ;;
    esac
    success "Added $install_dir to PATH"
elif [[ ":$PATH:" != *":$install_dir:"* ]]; then
    export PATH=$install_dir:$PATH
fi

echo ""
echo -e "${Green}  .___                 __         .__  .__                              ____    _______      _______${NC}"
echo -e "${Green}  |   | ____   _______/  |______  |  | |  |   ___________      ___  __ /_   |   \\   _  \\     \\   _  \\${NC}"
echo -e "${Green}  |   |/    \\ /  ___/\\   __\\__  \\ |  | |  | _/ __ \\_  __ \\     \\  \\/ /  |   |   /  /_\\  \\    /  /_\\  \\${NC}"
echo -e "${Green}  |   |   |  \\\\___ \\  |  |  / __ \\|  |_|  |\\_  ___/|  | \\/      \\   /   |   |   \\  \\_/   \\   \\  \\_/   \\${NC}"
echo -e "${Green}  |___|___|  /____  > |__| (____  /____/____/\\___  >__|          \\_/    |___| /\\ \\_____  / /\\ \\_____  /${NC}"
echo -e "${Green}           \\/     \\/            \\/               \\/                           \\/       \\/  \\/       \/${NC}"
echo ""
echo -e "${MUTED}  trashneurons-cli is installed and ready to use.${NC}"
echo -e "${MUTED}  Run 'trashneurons-cli' to start.${NC}"
echo ""
