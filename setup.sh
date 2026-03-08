#!/bin/bash

# OpenSearch Full-Text Search Setup Script
# This script sets up and tests the OpenSearch implementation

set -e

echo "=========================================="
echo "OpenSearch Full-Text Search Setup"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Check prerequisites
echo -e "${YELLOW}[1/6] Checking prerequisites...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker not found. Please install Docker.${NC}"
    exit 1
fi
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go not found. Please install Go 1.24+${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker and Go are installed${NC}"

# Step 2: Start OpenSearch cluster
echo ""
echo -e "${YELLOW}[2/6] Starting OpenSearch cluster...${NC}"
docker-compose down 2>/dev/null || true
docker-compose up -d
echo -e "${GREEN}✓ OpenSearch cluster started${NC}"
echo "  - OpenSearch API: http://localhost:9200"
echo "  - OpenSearch Dashboards: http://localhost:5601"

# Step 3: Wait for OpenSearch to be ready
echo ""
echo -e "${YELLOW}[3/6] Waiting for OpenSearch cluster to be ready...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:9200/_cluster/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OpenSearch is ready${NC}"
        break
    fi
    echo "  Waiting... ($i/30)"
    sleep 1
done

# Step 4: Build application
echo ""
echo -e "${YELLOW}[4/6] Building application...${NC}"
go mod tidy
go build ./...
echo -e "${GREEN}✓ Application built successfully${NC}"

# Step 5: Display information
echo ""
echo -e "${YELLOW}[5/6] Setup complete! Ready to start server.${NC}"
echo ""
echo "To start the application, run:"
echo "  go run main.go"
echo ""
echo "Then test with:"
echo "  # Upload a file"
echo "  curl -X POST http://localhost:8082/v1/files \\"
echo "    -F \"file=@test.txt\""
echo ""
echo "  # Search for files"
echo "  curl \"http://localhost:8082/v1/files?search=keyword\""
echo ""
echo "Monitor at:"
echo "  - OpenSearch Dashboards: http://localhost:5601"
echo "  - API: http://localhost:8082"
echo ""

# Step 6: Verify cluster health
echo -e "${YELLOW}[6/6] Verifying OpenSearch cluster health...${NC}"
CLUSTER_STATUS=$(curl -s http://localhost:9200/_cluster/health | grep -o '"status":"[^"]*' | cut -d'"' -f4)
if [ "$CLUSTER_STATUS" = "green" ]; then
    echo -e "${GREEN}✓ Cluster status: GREEN${NC}"
elif [ "$CLUSTER_STATUS" = "yellow" ]; then
    echo -e "${YELLOW}✓ Cluster status: YELLOW (normal for single machine)${NC}"
else
    echo -e "${RED}✗ Cluster status: $CLUSTER_STATUS${NC}"
fi

echo ""
echo -e "${GREEN}=========================================="
echo "Setup Complete!"
echo "==========================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Open a new terminal"
echo "  2. Run: cd /Users/mukate/GolandProjects/file_storage && go run main.go"
echo "  3. Test the API using the commands above"
echo "  4. Monitor at http://localhost:5601"
echo ""

