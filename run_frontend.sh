#!/usr/bin/env bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PROXY_FILE="frontend/proxy.conf.json"

echo -e "${GREEN}PSI Frontend - Test Script${NC}"
echo "============================="

# Load environment variables from .env file
if [ ! -f ".env" ]; then
    echo -e "${RED}Error: .env file not found!${NC}"
    echo -e "${YELLOW}Please create a .env file based on .env.example${NC}"
    echo -e "${YELLOW}You can copy it with: cp .env.example .env${NC}"
    exit 1
fi

echo -e "${GREEN}Loading environment variables from .env...${NC}"
set -a
source .env
set +a

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed${NC}"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo -e "${RED}Error: npm is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}Node.js version: $(node --version)${NC}"
echo -e "${GREEN}npm version: $(npm --version)${NC}"

# Install dependencies if node_modules doesn't exist
if [ ! -d "frontend/node_modules" ]; then
    echo -e "\n${GREEN}Installing dependencies...${NC}"
    (
        cd frontend
        npm install
    )
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to install dependencies${NC}"
        exit 1
    fi
    echo -e "${GREEN}Dependencies installed successfully${NC}"
else
    echo -e "\n${GREEN}Dependencies already installed${NC}"
fi

DEFAULT_URL="http://localhost:8080"

# Update proxy target to API_URL
echo -e "\n${GREEN}Updating proxy target to ${API_URL}...${NC}"
cp "$PROXY_FILE" "${PROXY_FILE}.bak"
sed -i "s|${DEFAULT_URL}|${API_URL}|g" "$PROXY_FILE"

# Restore proxy config on exit
cleanup() {
    echo -e "\n${YELLOW}Restoring proxy configuration...${NC}"
    mv "${PROXY_FILE}.bak" "$PROXY_FILE"
    echo -e "${GREEN}Proxy configuration restored${NC}"
    exit 0
}
trap cleanup SIGINT SIGTERM

# Start Angular dev server
echo -e "${GREEN}Starting Angular development server...${NC}"
echo -e "${YELLOW}The frontend will be available at: ${FRONTEND_URL}${NC}"
echo -e "${YELLOW}Proxying API requests to: ${API_URL}${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}\n"

(cd frontend && npm start)
