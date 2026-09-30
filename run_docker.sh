#!/usr/bin/env bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}PSI Project - Docker Execution Script${NC}"
echo "========================================"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    exit 1
fi

# Check if Docker daemon is running and user has access
DOCKER_INFO_ERR=$(docker info 2>&1 >/dev/null)
if [ $? -ne 0 ]; then
    if echo "$DOCKER_INFO_ERR" | grep -qi "permission denied"; then
        echo -e "${RED}Error: Docker daemon is running but you don't have permission to access it${NC}"
        echo -e "${YELLOW}Your user is probably not in the 'docker' group${NC}"
        echo -e "${YELLOW}You can fix this by running: sudo usermod -aG docker \$USER${NC}"
        echo -e "${YELLOW}Then log out and log back in, or run: newgrp docker${NC}"
        echo -e "${YELLOW}Alternatively, you can run this script with sudo${NC}"
    else
        echo -e "${RED}Error: Docker daemon is not running${NC}"
        echo -e "${YELLOW}Please start Docker and try again${NC}"
    fi
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    exit 1
fi

# Determine docker compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Stopping Docker containers...${NC}"
    $DOCKER_COMPOSE down
    echo -e "${GREEN}Containers stopped${NC}"
    exit 0
}

# Trap Ctrl+C and cleanup
trap cleanup SIGINT SIGTERM

echo -e "\n${GREEN}Starting services with Docker Compose...${NC}"
$DOCKER_COMPOSE up --build

# If docker-compose exits, run cleanup
cleanup
