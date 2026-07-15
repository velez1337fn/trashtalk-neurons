#!/usr/bin/env bash
set -euo pipefail

APP=trashneurons-cli
MUTED=$'\033[0;2m'
RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
BLUE=$'\033[0;34m'
ORANGE=$'\033[38;5;214m'
RESET=$'\033[0m'

if [[ -t 1 ]]; then
    Bold_White=$'\033[1m'
    Bold_Green=$'\033[1;32m'
else
    Bold_White=''
    Bold_Green=''
fi

error() {
    echo -e "${RED}$@ ${RESET}" >&2
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

border_top="${BLUE}┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────${RESET}"
border_bot="${BLUE}└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────${RESET}"
border_left="${BLUE}│${RESET}"
border_right="${BLUE}│${RESET}"
border_corner="${BLUE}├──────────────────────┐${RESET}${BLUE}                                                                                 ${BLUE}├────────────────────────────────────────────────────────${RESET}"
border_corner_bot="${BLUE}└──────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘${RESET}"

echo -e "$border_top"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left} ${GREEN}>${RESET} allocating path where app/cli been installed...                                                                    ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left}                 ${GREEN}found! path is:${RESET} ${install_dir}                                                                            ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left} ${GREEN}>${RESET} starting download via curl...                                                                                       ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"

tmp_dir="${TMPDIR:-/tmp}/trashneurons_install_$$"
mkdir -p "$tmp_dir"

curl --fail --location --progress-bar --output "$tmp_dir/$filename" "$url" ||
    error "Failed to download from $url"

tar -xzf "$tmp_dir/$filename" -C "$tmp_dir"

mv "$tmp_dir/$APP" "$exe"
chmod +x "$exe"
rm -rf "$tmp_dir"

echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left}    ${GREEN}installing your app/cli...${RESET}                                                                                          ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left}  ${GREEN}finished install!${RESET} all files located in ${install_dir}                                                                   ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "${border_left}  ${MUTED}for use:${RESET} type \"trashneurons-cli\"                                                                                    ${border_right}"
echo -e "${border_left}                                                                                                                        ${border_right}"
echo -e "$border_corner"
echo -e "${border_left} ${RED}PRESS CTRL+C TO CLOSE INSTALLER${RESET}   ${border_right}"
echo -e "$border_corner_bot"

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
        *) echo "" >> "$config_file"; echo "# trashneurons" >> "$config_file"; echo "export PATH=$install_dir:\$PATH" >> "$config_file" ;;
    esac
    echo -e "${GREEN}Added $install_dir to PATH${RESET}"
elif [[ ":$PATH:" != *":$install_dir:"* ]]; then
    export PATH=$install_dir:$PATH
fi

echo ""
echo -e "${GREEN}  .___                 __         .__  .__                              ____    _______      _______${RESET}"
echo -e "${GREEN}  |   | ____   _______/  |______  |  | |  |   ___________      ___  __ /_   |   \\   _  \\     \\   _  \\${RESET}"
echo -e "${GREEN}  |   |/    \\ /  ___/\\   __\\__  \\ |  | |  | _/ __ \\_  __ \\     \\  \\/ /  |   |   /  /_\\  \\    /  /_\\  \\${RESET}"
echo -e "${GREEN}  |   |   |  \\\\___ \\  |  |  / __ \\|  |_|  |\\_  ___/|  | \\/      \\   /   |   |   \\  \\_/   \\   \\  \\_/   \\${RESET}"
echo -e "${GREEN}  |___|___|  /____  > |__| (____  /____/____/\\___  >__|          \\_/    |___| /\\ \\_____  / /\\ \\_____  /${RESET}"
echo -e "${GREEN}           \\/     \\/            \\/               \\/                           \\/       \\/  \\/       \/${RESET}"
echo ""
echo -e "${MUTED}  trashneurons-cli is installed and ready to use.${RESET}"
echo -e "${MUTED}  Run 'trashneurons-cli' to start.${RESET}"
echo ""
