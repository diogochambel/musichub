#!/usr/bin/env bash

# Format Go backend code with gofmt
# Usage: ./scripts/format_backend.sh [--check]

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

CHECK_MODE=false
if [ "$1" = "--check" ]; then
    CHECK_MODE=true
fi

if ! command -v gofmt &> /dev/null; then
    echo -e "${RED}Error: gofmt is not installed${NC}"
    echo -e "${YELLOW}Install Go: https://go.dev/dl/${NC}"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="${ROOT_DIR}/backend"

if [ ! -d "$BACKEND_DIR" ]; then
    echo -e "${RED}Error: backend/ directory not found at ${BACKEND_DIR}${NC}"
    exit 1
fi

if [ "$CHECK_MODE" = true ]; then
    echo -e "${YELLOW}Checking Go formatting...${NC}"
    UNFORMATTED=$(gofmt -l "$BACKEND_DIR")
    if [ -z "$UNFORMATTED" ]; then
        echo -e "${GREEN}All Go files are properly formatted${NC}"
        exit 0
    else
        echo -e "${RED}The following files need formatting:${NC}"
        echo "$UNFORMATTED"
        exit 1
    fi
else
    echo -e "${GREEN}Formatting Go files...${NC}"
    gofmt -w "$BACKEND_DIR"
    echo -e "${GREEN}Done${NC}"
fi
