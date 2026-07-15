#!/usr/bin/env bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

if [[ "$(uname)" != "Linux" ]]; then
    echo -e "${RED}Ошибка: только Linux.${NC}"
    exit 1
fi

clear

cat << "EOF"
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────┬───────────────────────┐
│   .___                 __         .__  .__                              ____    _______      _______     │telegram - @trashtalkAI│
│   |   | ____   _______/  |______  |  | |  |   ___________      ___  __ /_   |   \   _  \     \   _  \    │                       │
│   |   |/    \ /  ___/\   __\__  \ |  | |  | _/ __ \_  __ \     \  \/ /  |   |   /  /_\  \    /  /_\  \   ├───────────────────────┤
│   |   |   |  \\___ \  |  |  / __ \|  |_|  |_\  ___/|  | \/      \   /   |   |   \  \_/   \   \  \_/   \  │                       │
│   |___|___|  /____  > |__| (____  /____/____/\___  >__|          \_/    |___| /\ \_____  / /\ \_____  /  │site: трештолк.рф      │
│            \/     \/            \/               \/                           \/       \/  \/       \/   │                       │
├─────────────────────┐─────────────────────────────────────────────────────────────────┌──────────────────────────────────────────┤
│                     │                    _           _                                │                                          │
│  welcome, "$USER" ! │           ___  ___| | ___  ___| |_                              │                                          │
│                     │          / __|/ _ \ |/ _ \/ __| __|                             │                                          │
├─────────────────────┤          \__ \  __/ |  __/ (__| |_                              │    detected sys: "$(uname -s -r)"        │
│                     │          |___/\___|_|\___|\___|\__|   _                         │                                          │
│                     │            __ _ _ __  _   _| |_| |__ (_)_ __   __ _             │    installer status: ready               │
│ "matrix in this box"│           / _` | '_ \| | | | __| '_ \| | '_ \ / _` |            │                                          │
│                     │      ─   | (_| | | | | |_| | |_| | | | | | | | (_| |            │    waiting for user input...             │
│                     │           \__,_|_| |_|\__, |\__|_| |_|_|_| |_|\__, |            │                                          │
│                     │                       |___/                   |___/             │                                          │
│                     │┌───────────────────┬───────────────────┬───────────────────────┐├──────────────────────────────────────────┤
│                     ││  1. desktop app   │  2. CLI           │     3. both           ││                                          │
│                     │└───────────────────┴───────────────────┴───────────────────────┘│  are ya winning son?     ***   ***       │
│                     │                                                                 │    ─┬────┐             ***** *****       │
│                     │                                                                 │     └─┬──┘              *********        │
│                     │                                                                 │   ┌───┼──┐               *******         │
│                     │           ┌──────────────────────────────────────┐              │   │   │  │                *****          │
│                     │           │ user input                           │              │     ┌─┼─┐                  ***           │
│                     │           └──────────────────────────────────────┘              │     │   │                   *            │
└─────────────────────┴─────────────────────────────────────────────────────────────────└─────┴───┴────────────────────────────────┘
EOF

echo -e "\n${YELLOW}Выберите вариант:${NC}"
echo "1) Desktop (в разработке)"
echo "2) CLI"
echo "3) Оба (Desktop - заглушка)"
read -p "Введите номер: " choice

case $choice in
    1) echo -e "${RED}Desktop пока не реализован.${NC}"; exit 0 ;;
    2) INSTALL_CLI=true ;;
    3) INSTALL_CLI=true; echo -e "${YELLOW}Будет установлен только CLI.${NC}" ;;
    *) echo -e "${RED}Неверный выбор.${NC}"; exit 1 ;;
esac

if [ "$INSTALL_CLI" = true ]; then
    echo -e "\n${GREEN}Установка CLI...${NC}"
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"

    echo -e "${BLUE}Проверка соединения с GitHub...${NC}"
    if ! curl -sSf --connect-timeout 5 https://github.com > /dev/null; then
        echo -e "${RED}Не удаётся подключиться к GitHub.${NC}"
        echo -e "${YELLOW}Попробуйте: echo 'nameserver 1.1.1.1' | sudo tee /etc/resolv.conf${NC}"
        exit 1
    fi

    # Проверяем, установлен ли уже бинарник
    if [ -f "$INSTALL_DIR/trashneurons-cli" ]; then
        echo -e "${YELLOW}Обнаружена предыдущая установка CLI. Будет выполнена переустановка.${NC}"
        rm -f "$INSTALL_DIR/trashneurons-cli"
    fi

    echo -e "${BLUE}Скачивание бинарника...${NC}"
    if ! curl -L -o /tmp/trashneurons-cli https://github.com/velez1337fn/trashtalk-neurons/releases/latest/download/trashneurons-cli --progress-bar; then
        echo -e "${RED}Ошибка скачивания.${NC}"
        exit 1
    fi

    chmod +x /tmp/trashneurons-cli
    mv /tmp/trashneurons-cli "$INSTALL_DIR/trashneurons-cli"

    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo -e "${YELLOW}Добавляем ~/.local/bin в PATH.${NC}"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
        export PATH="$HOME/.local/bin:$PATH"
    fi

    mkdir -p "$HOME/.config/trashneurons"

    echo -e "\n${GREEN}Установка завершена!${NC}"
    echo -e "Теперь запускайте: ${YELLOW}trashneurons-cli${NC}"
    echo -e "Если команда не найдена, перезапустите терминал или выполните: ${YELLOW}source ~/.bashrc${NC}"
fi

echo -e "\n${GREEN}Нажмите Enter для выхода.${NC}"
read
