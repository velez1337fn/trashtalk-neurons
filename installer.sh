#!/usr/bin/env bash
set -euo pipefail

APP=trashneurons-cli

error() {
    echo "error: $@" >&2
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

echo "> allocating path where app/cli been installed..."
echo "  found! path is: $install_dir"
echo ""
echo "> starting download via curl..."

tmp_dir="${TMPDIR:-/tmp}/trashneurons_install_$$"
mkdir -p "$tmp_dir"

curl --fail --location --progress-bar --output "$tmp_dir/$filename" "$url" ||
    error "Failed to download from $url"

echo "  downloaded!"
echo ""
echo "    installing your app/cli..."

tar -xzf "$tmp_dir/$filename" -C "$tmp_dir"

mv "$tmp_dir/$APP" "$exe"
chmod +x "$exe"
rm -rf "$tmp_dir"

echo ""
echo "  finished install! all files located in $install_dir"
echo ""
echo "  for use: type \"trashneurons-cli\""
echo ""
echo "  PRESS ENTER TO CLOSE INSTALLER"
read -r

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
    echo "Added $install_dir to PATH"
elif [[ ":$PATH:" != *":$install_dir:"* ]]; then
    export PATH=$install_dir:$PATH
fi

echo ""
echo "  .___                 __         .__  .__                              ____    _______      _______"
echo "  |   | ____   _______/  |______  |  | |  |   ___________      ___  __ /_   |   \\   _  \\     \\   _  \\"
echo "  |   |/    \\ /  ___/\\   __\\__  \\ |  | |  | _/ __ \\_  __ \\     \\  \\/ /  |   |   /  /_\\  \\    /  /_\\  \\"
echo "  |   |   |  \\\\___ \\  |  |  / __ \\|  |_|  |\\_  ___/|  | \\/      \\   /   |   |   \\  \\_/   \\   \\  \\_/   \\"
echo "  |___|___|  /____  > |__| (____  /____/____/\\___  >__|          \\_/    |___| /\\ \\_____  / /\\ \\_____  /"
echo "           \\/     \\/            \\/               \\/                           \\/       \\/  \\/       \\/"
echo ""
echo "  trashneurons-cli is installed and ready to use."
echo "  Run 'trashneurons-cli' to start."
echo ""
