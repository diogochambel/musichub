#!/usr/bin/env bash

# Format all project code (Go backend + Angular frontend)
# Usage: ./scripts/format_all.sh [--check]

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

CHECK_FLAG=""
if [ "$1" = "--check" ]; then
    CHECK_FLAG="--check"
fi

echo -e "${GREEN}Formatting all project code${NC}"
echo "================================"

FAIL=0

echo ""
echo "==> Backend (Go)"
"${SCRIPT_DIR}/format_backend.sh" $CHECK_FLAG
if [ $? -ne 0 ]; then
    FAIL=1
fi

echo ""
echo "==> Frontend (Angular/TypeScript)"
"${SCRIPT_DIR}/format_frontend.sh" $CHECK_FLAG
if [ $? -ne 0 ]; then
    FAIL=1
fi

echo ""
echo "================================"
if [ $FAIL -ne 0 ]; then
    echo -e "${RED}Some formatting checks failed${NC}"
    exit 1
else
    echo -e "${GREEN}All code formatted successfully${NC}"
    exit 0
fi