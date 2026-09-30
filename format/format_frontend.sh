#!/usr/bin/env bash

# Format frontend code with Prettier
# Usage: ./scripts/format_frontend.sh [--check]

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

CHECK_MODE=false
if [ "$1" = "--check" ]; then
    CHECK_MODE=true
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
FRONTEND_DIR="${ROOT_DIR}/frontend"

if ! command -v npm &> /dev/null; then
    echo -e "${RED}Error: npm is not installed${NC}"
    echo -e "${YELLOW}Install Node.js: https://nodejs.org/${NC}"
    exit 1
fi

if [ ! -d "$FRONTEND_DIR" ]; then
    echo -e "${RED}Error: frontend/ directory not found at ${FRONTEND_DIR}${NC}"
    exit 1
fi

cd "$FRONTEND_DIR"

if ! npx prettier --version &> /dev/null; then
    echo -e "${YELLOW}Prettier not found, installing as devDependency...${NC}"
    npm install --save-dev prettier
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to install Prettier${NC}"
        exit 1
    fi
    echo -e "${GREEN}Prettier installed${NC}"
fi

PRETTIER_CMD="npx prettier"
GLOB_PATTERNS=("src/**/*.{ts,html,css,scss,json}")

if [ "$CHECK_MODE" = true ]; then
    echo -e "${YELLOW}Checking frontend formatting...${NC}"
    $PRETTIER_CMD --check "${GLOB_PATTERNS[@]}"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}All frontend files are properly formatted${NC}"
        exit 0
    else
        echo -e "${RED}Some files need formatting${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}Formatting frontend files...${NC}"
    $PRETTIER_CMD --write "${GLOB_PATTERNS[@]}"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Prettier failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}Done${NC}"
fi