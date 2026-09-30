#!/usr/bin/env bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PROXY_FILE="frontend/proxy.conf.json"
DEFAULT_URL="http://localhost:8080"

echo -e "${GREEN}PSI Project - Local Execution Script${NC}"
echo "========================================"

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

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed${NC}"
    exit 1
fi

# Check if MongoDB is running
if ! command -v mongosh &> /dev/null && ! command -v mongo &> /dev/null; then
    echo -e "${YELLOW}Warning: MongoDB client not found. Make sure MongoDB is running on localhost:27017${NC}"
fi

# Create bin directory if it doesn't exist
mkdir -p bin

# Build backend
echo -e "\n${GREEN}Building backend...${NC}"
(
    cd backend
    go build -o ../bin/server ./cmd/server
    go build -o ../bin/populate ./cmd/populate
)
if [ $? -ne 0 ]; then
    echo -e "${RED}Backend build failed${NC}"
    exit 1
fi
echo -e "${GREEN}Backend built successfully -> ./bin/server${NC}"

# Populate database
echo -e "\n${GREEN}Populating database...${NC}"
./bin/populate

# Install frontend dependencies if needed
if [ ! -d "frontend/node_modules" ]; then
    echo -e "\n${GREEN}Installing frontend dependencies...${NC}"
    (
        cd frontend
        npm install
    )
    if [ $? -ne 0 ]; then
        echo -e "${RED}Frontend dependency installation failed${NC}"
        exit 1
    fi
fi

# Start backend in background
echo -e "\n${GREEN}Starting backend on port ${SERVER_PORT}...${NC}"
./bin/server &
BACKEND_PID=$!
echo -e "${GREEN}Backend started with PID: $BACKEND_PID${NC}"

# Wait a moment for backend to start
sleep 2

# Update proxy target to API_URL
echo -e "\n${GREEN}Updating proxy target to ${API_URL}...${NC}"
cp "$PROXY_FILE" "${PROXY_FILE}.bak"
sed -i "s|${DEFAULT_URL}|${API_URL}|g" "$PROXY_FILE"

# Start frontend
echo -e "\n${GREEN}Starting frontend on port 4200...${NC}"
(cd frontend && npm start) &
FRONTEND_PID=$!
echo -e "${GREEN}Frontend started with PID: $FRONTEND_PID${NC}"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Stopping services...${NC}"
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    # Kill any remaining node processes from Angular
    pkill -P $FRONTEND_PID 2>/dev/null
    # Restore proxy configuration
    if [ -f "${PROXY_FILE}.bak" ]; then
        echo -e "${YELLOW}Restoring proxy configuration...${NC}"
        mv "${PROXY_FILE}.bak" "$PROXY_FILE"
        echo -e "${GREEN}Proxy configuration restored${NC}"
    fi
    echo -e "${GREEN}Services stopped${NC}"
    exit 0
}

# Trap Ctrl+C and cleanup
trap cleanup SIGINT SIGTERM

echo -e "\n${GREEN}All services are running!${NC}"
echo -e "${GREEN}Backend:  http://localhost:${SERVER_PORT}${NC}"
echo -e "${GREEN}Frontend: ${FRONTEND_URL}${NC}"
echo -e "\n${YELLOW}Press Ctrl+C to stop all services${NC}\n"

# Wait for user interrupt
wait
