#!/usr/bin/env bash
set -euo pipefail
APP=trashneurons-cli

MUTED='\033[0;2m'
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
ORANGE='\033[38;5;214m'
NC='\033[0m'

usage() {
    cat <<EOF
TrashNeurons Installer

Usage: installer.sh [options]

Options:
    -h, --help              Display this help message
    -v, --version <version> Install a specific version (e.g., v1.0.0)

Examples:
    curl -fsSL https://raw.githubusercontent.com/velez1337fn/trashtalk-neurons/main/installer.sh | bash
    curl -fsSL https://raw.githubusercontent.com/velez1337fn/trashtalk-neurons/main/installer.sh | bash -s -- -v v1.0.0
EOF
}

requested_version=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--version)
            if [[ -n "${2:-}" ]]; then
                requested_version="$2"
                shift 2
            else
                echo -e "${RED}Error: --version requires a version argument${NC}"
                exit 1
            fi
            ;;
        *)
            echo -e "${ORANGE}Warning: Unknown option '$1'${NC}" >&2
            shift
            ;;
    esac
done

INSTALL_DIR="$HOME/.trashneurons/bin"
mkdir -p "$INSTALL_DIR"

raw_os=$(uname -s)
os=$(echo "$raw_os" | tr '[:upper:]' '[:lower:]')
case "$raw_os" in
  Darwin*) os="darwin" ;;
  Linux*) os="linux" ;;
  MINGW*|MSYS*|CYGWIN*) os="windows" ;;
esac

arch=$(uname -m)
if [[ "$arch" == "aarch64" ]]; then
  arch="arm64"
fi
if [[ "$arch" == "x86_64" ]]; then
  arch="amd64"
fi

combo="$os-$arch"
case "$combo" in
  linux-amd64|linux-arm64|darwin-amd64|darwin-arm64)
    ;;
  *)
    echo -e "${RED}Unsupported OS/Arch: $os/$arch${NC}"
    exit 1
    ;;
esac

archive_ext=".tar.gz"
if [ "$os" = "darwin" ]; then
  archive_ext=".zip"
fi

if [ "$os" = "linux" ]; then
    if ! command -v tar >/dev/null 2>&1; then
         echo -e "${RED}Error: 'tar' is required but not installed.${NC}"
         exit 1
    fi
fi

if [ -z "$requested_version" ]; then
    tag="v1.0.0"
else
    tag="$requested_version"
fi

filename="${APP}-${os}-${arch}.tar.gz"
url="https://github.com/velez1337fn/trashtalk-neurons/releases/download/${tag}/${filename}"

print_message() {
    local level=$1
    local message=$2
    local color=""

    case $level in
        info) color="${NC}" ;;
        warning) color="${NC}" ;;
        error) color="${RED}" ;;
    esac

    echo -e "${color}${message}${NC}"
}

print_message info "${BLUE}┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC} ${GREEN}>${NC} allocating path where app/cli been installed...                                                                    ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                 ${GREEN}found! path is:${NC} ${INSTALL_DIR}                                                                            ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC} ${GREEN}>${NC} starting download via curl...                                                                                       ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"

tmp_dir="${TMPDIR:-/tmp}/trashneurons_install_$$"
mkdir -p "$tmp_dir"

if [[ "$os" == "darwin" ]] || ! [ -t 2 ] || ! download_with_progress "$url" "$tmp_dir/$filename"; then
    curl -# -L -o "$tmp_dir/$filename" "$url"
fi

tar -xzf "$tmp_dir/$filename" -C "$tmp_dir"

mv "$tmp_dir/${APP}" "$INSTALL_DIR/${APP}"
chmod 755 "${INSTALL_DIR}/${APP}"
rm -rf "$tmp_dir"

print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}    ${GREEN}installing your app/cli...${NC}                                                                                          ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}  ${GREEN}finished install!${NC} all files located in ${INSTALL_DIR}                                                                   ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}  ${MUTED}for use:${NC} type \"trashneurons-cli\"                                                                                    ${BLUE}│${NC}"
print_message info "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
print_message info "${BLUE}├──────────────────────┐${NC}                                                                                 ${BLUE}├────────────────────────────────────────────────────────┤${NC}"
print_message info "${BLUE}│${NC} ${RED}PRESS CTRL+C TO CLOSE INSTALLER${NC}   ${BLUE}│${NC}                                                                                 ${BLUE}│${NC}"
print_message info "${BLUE}└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘${NC}"

XDG_CONFIG_HOME=${XDG_CONFIG_HOME:-$HOME/.config}
current_shell=$(basename "$SHELL" 2>/dev/null || echo "bash")

case $current_shell in
    fish)
        config_file="$HOME/.config/fish/config.fish"
        ;;
    zsh)
        config_file="${ZDOTDIR:-$HOME}/.zshrc"
        ;;
    bash)
        config_file="$HOME/.bashrc"
        ;;
    *)
        config_file="$HOME/.bashrc"
        ;;
esac

if [[ -f "$config_file" ]] && [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    case $current_shell in
        fish)
            echo "fish_add_path $INSTALL_DIR" >> "$config_file"
            ;;
        *)
            echo "" >> "$config_file"
            echo "# trashneurons" >> "$config_file"
            echo "export PATH=$INSTALL_DIR:\$PATH" >> "$config_file"
            ;;
    esac
    print_message info "${GREEN}Added ${INSTALL_DIR} to PATH in ${config_file}${NC}"
elif [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    export PATH=$INSTALL_DIR:$PATH
    print_message warning "Manually add to your shell config:"
    print_message info "  export PATH=$INSTALL_DIR:\$PATH"
fi

echo ""
echo -e "${GREEN}  .___                 __         .__  .__                              ____    _______      _______${NC}"
echo -e "${GREEN}  |   | ____   _______/  |______  |  | |  |   ___________      ___  __ /_   |   \\   _  \\     \\   _  \\${NC}"
echo -e "${GREEN}  |   |/    \\ /  ___/\\   __\\__  \\ |  | |  | _/ __ \\_  __ \\     \\  \\/ /  |   |   /  /_\\  \\    /  /_\\  \\${NC}"
echo -e "${GREEN}  |   |   |  \\\\___ \\  |  |  / __ \\|  |_|  |\\_  ___/|  | \\/      \\   /   |   |   \\  \\_/   \\   \\  \\_/   \\${NC}"
echo -e "${GREEN}  |___|___|  /____  > |__| (____  /____/____/\\___  >__|          \\_/    |___| /\\ \\_____  / /\\ \\_____  /${NC}"
echo -e "${GREEN}           \\/     \\/            \\/               \\/                           \\/       \\/  \\/       \/${NC}"
echo ""
echo -e "${MUTED}  trashneurons-cli is installed and ready to use.${NC}"
echo -e "${MUTED}  Run 'trashneurons-cli' to start.${NC}"
echo ""

download_with_progress() {
    local url="$1"
    local output="$2"

    if [ -t 2 ]; then
        exec 4>&2
    else
        exec 4>/dev/null
    fi

    local tmp_dir=${TMPDIR:-/tmp}
    local basename="${tmp_dir}/trashneurons_install_$$"
    local tracefile="${basename}.trace"

    rm -f "$tracefile"
    mkfifo "$tracefile"

    printf "\033[?25l" >&4

    trap "trap - RETURN; rm -f \"$tracefile\"; printf '\033[?25h' >&4; exec 4>&-" RETURN

    (
        curl --trace-ascii "$tracefile" -s -L -o "$output" "$url"
    ) &
    local curl_pid=$!

    sed -n -e 'y/ACDEGHLNORTV/acdeghlnortv/' -e '/^0000: content-length:/p' -e '/^<= recv data/p' "$tracefile" | \
    {
        local length=0
        local bytes=0

        while IFS=" " read -r -a line; do
            [ "${#line[@]}" -lt 2 ] && continue
            local tag="${line[0]} ${line[1]}"

            if [ "$tag" = "0000: content-length:" ]; then
                length="${line[2]}"
                length=$(echo "$length" | tr -d '\r')
                bytes=0
            elif [ "$tag" = "<= recv" ]; then
                local size="${line[3]}"
                bytes=$(( bytes + size ))
                if [ "$length" -gt 0 ]; then
                    local width=50
                    local percent=$(( bytes * 100 / length ))
                    [ "$percent" -gt 100 ] && percent=100
                    local on=$(( percent * width / 100 ))
                    local off=$(( width - on ))
                    local filled=$(printf "%*s" "$on" "")
                    filled=${filled// /■}
                    local empty=$(printf "%*s" "$off" "")
                    empty=${empty// /･}
                    printf "\r${ORANGE}%s%s %3d%%${NC}" "$filled" "$empty" "$percent" >&4
                fi
            fi
        done
    }

    wait $curl_pid
    local ret=$?
    echo "" >&4
    return $ret
}
