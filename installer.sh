#!/bin/bash

set -e

API_KEY="sk-L2HKJL0VysIiDI9MiibyXyppApPb6Z7FQFYXo7qKs1STf18L"
REPO_URL="https://github.com/velez1337fn/trashtalk-neurons"
INSTALL_DIR="$HOME/.trashneurons"
BIN_NAME="trashneurons-cli"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_logo() {
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
}

print_side_info() {
    echo -e "${YELLOW}telegram - @trashtalkAI${NC}"
    echo -e "${YELLOW}site: трештолк.рф${NC}"
}

detect_system() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch
    arch=$(uname -m)
    
    case "$arch" in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        armv7l) arch="arm" ;;
    esac
    
    echo "${os}-${arch}"
}

show_welcome_screen() {
    clear
    print_logo
    
    local user
    user=$(whoami)
    local sys_info
    sys_info=$(detect_system)
    
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
    read -r choice
    
    case "$choice" in
        1) install_cli ;;
        2) exit 0 ;;
        *) echo -e "${RED}Invalid option. Exiting.${NC}"; exit 1 ;;
    esac
}

install_cli() {
    clear
    echo -e "${BLUE}┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
    
    echo -e "${BLUE}│${NC} ${YELLOW}>${NC} allocating path where app/cli been installed...                                                                    ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
    
    mkdir -p "$INSTALL_DIR"
    echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC}                 ${GREEN}found! path is:${NC} ${INSTALL_DIR}                                                                            ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC} ${YELLOW}>${NC} starting download via curl...                                                                                       ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
    
    local download_dir
    download_dir=$(mktemp -d)
    
    cd "$download_dir"
    git clone --quiet "$REPO_URL" trashtalk-neurons 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
        
        local total_size
        total_size=$(du -sb "$download_dir/trashtalk-neurons/cli" | cut -f1)
        local copied=0
        
        echo -e "${BLUE}│${NC} ${GREEN}copying files...${NC}                                                                                                     ${BLUE}│${NC}"
        
        mkdir -p "$INSTALL_DIR/bin"
        cp -r "$download_dir/trashtalk-neurons/cli/"* "$INSTALL_DIR/bin/" 2>/dev/null || true
        cp -r "$download_dir/trashtalk-neurons/assets/"* "$INSTALL_DIR/assets/" 2>/dev/null || true
        
        echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC} ${GREEN}finished download!${NC}                                                                                                    ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC}    ${YELLOW}installing your app/cli...${NC}                                                                                          ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
        
        chmod +x "$INSTALL_DIR/bin/trashneurons-cli" 2>/dev/null || true
        
        if [ ! -f "$INSTALL_DIR/bin/go.mod" ]; then
            go build -o "$INSTALL_DIR/bin/trashneurons-cli" "$INSTALL_DIR/bin/" 2>/dev/null || true
        fi
        
        if [ ! -L "/usr/local/bin/$BIN_NAME" ] && [ ! -f "/usr/local/bin/$BIN_NAME" ]; then
            sudo ln -sf "$INSTALL_DIR/bin/trashneurons-cli" "/usr/local/bin/$BIN_NAME" 2>/dev/null || true
        fi
        
        echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC}  ${GREEN}finished install!${NC} all files located in ${INSTALL_DIR}                                                                   ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC}  ${YELLOW}for use:${NC} type \"trashneurons-cli\"                                                                                    ${BLUE}│${NC}"
        echo -e "${BLUE}│${NC}                                                                                                                        ${BLUE}│${NC}"
        echo -e "${BLUE}├──────────────────────┐${NC}                                                                                 ${BLUE}├────────────────────────────────────────────────────────┤${NC}"
        echo -e "${BLUE}│${NC} ${RED}PRESS CTRL+C TO CLOSE INSTALLER${NC}   ${BLUE}│${NC}                                                                                 ${BLUE}│${NC}"
        echo -e "${BLUE}└──────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘${NC}"
    else
        echo -e "${BLUE}│${NC} ${RED}ERROR: Failed to clone repository${NC}                                                                               ${BLUE}│${NC}"
        echo -e "${BLUE}└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘${NC}"
        rm -rf "$download_dir"
        exit 1
    fi
    
    rm -rf "$download_dir"
    
    echo ""
    echo -e "${GREEN}Press Enter to exit...${NC}"
    read -r
}

show_welcome_screen
