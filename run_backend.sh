#!/usr/bin/env bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}PSI Backend - Test Script${NC}"
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

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Check if MongoDB is running
echo -e "\n${YELLOW}Checking MongoDB connection...${NC}"
MONGODB_HOST=$(echo "${MONGODB_URI}" | sed 's|mongodb://||' | cut -d'/' -f1)
if ! command -v mongosh &> /dev/null && ! command -v mongo &> /dev/null; then
    echo -e "${YELLOW}Warning: MongoDB client not found. Make sure MongoDB is running at ${MONGODB_HOST}${NC}"
else
    # Try to ping MongoDB
    if command -v mongosh &> /dev/null; then
        if ! mongosh --eval "db.adminCommand('ping')" --quiet "${MONGODB_URI}/test" &> /dev/null; then
            echo -e "${RED}Error: Cannot connect to MongoDB at ${MONGODB_URI}${NC}"
            echo -e "${YELLOW}Please start MongoDB first or run: docker run -d -p 27017:27017 mongo${NC}"
            exit 1
        fi
    fi
fi

echo -e "${GREEN}MongoDB is accessible${NC}"

# Create bin directory if it doesn't exist
mkdir -p bin

# Build backend
echo -e "\n${GREEN}Building backend...${NC}"
(
    cd backend || exit
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

# Start backend
echo -e "\n${GREEN}Starting backend server...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}\n"

./bin/server
