#!/usr/bin/env bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}PSI MongoDB - Docker Runner${NC}"
echo "============================"

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

echo -e "${GREEN}Docker is installed${NC}"

# Extract host and port from MONGODB_URI
# Default to localhost:27017 if not set
MONGODB_HOST=$(echo "${MONGODB_URI:-mongodb://localhost:27017}" | sed 's|mongodb://||' | cut -d'/' -f1)
MONGODB_PORT=$(echo "${MONGODB_HOST}" | cut -d':' -f2)
MONGODB_HOST=$(echo "${MONGODB_HOST}" | cut -d':' -f1)

# Default to 27017 if no port specified
if [ -z "$MONGODB_PORT" ]; then
    MONGODB_PORT=27017
fi

echo -e "\n${GREEN}MongoDB Configuration:${NC}"
echo -e "  Host: ${MONGODB_HOST}"
echo -e "  Port: ${MONGODB_PORT}"
echo -e "  Image: mongo:7.0"

# Check if MongoDB is already running via Docker
if docker ps | grep -q "psi-mongodb"; then
    echo -e "\n${YELLOW}MongoDB container 'psi-mongodb' is already running${NC}"
    echo -e "${YELLOW}You can stop it with: docker stop psi-mongodb${NC}"
    exit 0
fi

# Check if port is already in use
if lsof -Pi :${MONGODB_PORT} -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "\n${YELLOW}Warning: Port ${MONGODB_PORT} is already in use${NC}"
    echo -e "${YELLOW}Make sure another MongoDB instance isn't running locally${NC}"
fi

# Start MongoDB container
echo -e "\n${GREEN}Starting MongoDB 7.0 container...${NC}"
echo -e "${YELLOW}Container name: psi-mongodb${NC}"
echo -e "${YELLOW}Data volume: psi-mongodb-data${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}\n"

# Create volume if it doesn't exist
docker volume create psi-mongodb-data 2>/dev/null || true

# Run MongoDB container with auto-remove
docker run \
    --name psi-mongodb \
    --rm \
    -p ${MONGODB_PORT}:27017 \
    -v psi-mongodb-data:/data/db \
    mongo:7.0

# Cleanup message after container stops
echo -e "\n${GREEN}MongoDB container stopped${NC}"
echo -e "${YELLOW}Data is persisted in Docker volume 'psi-mongodb-data'${NC}"
