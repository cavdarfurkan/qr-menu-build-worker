#!/bin/bash
set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ENV_FILE="/etc/qr-menu-worker.env"

echo -e "${GREEN}=== QR Menu Build Worker Environment Setup ===${NC}"
echo ""

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "Error: This script must be run as root (use sudo)"
   exit 1
fi

# Check if environment file already exists
if [ -f "${ENV_FILE}" ]; then
    echo -e "${YELLOW}Warning: ${ENV_FILE} already exists!${NC}"
    read -p "Do you want to overwrite it? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted. Existing file not modified."
        exit 0
    fi
fi

# Prompt for values
echo "Please provide the following configuration values:"
echo ""

read -p "Redis endpoint (e.g., xxx.cache.amazonaws.com): " REDIS_HOST
REDIS_ADDR="${REDIS_HOST}:6379"

read -p "Queue key [queue:build:main]: " QUEUE_KEY
QUEUE_KEY=${QUEUE_KEY:-queue:build:main}

read -p "AWS Region [eu-central-1]: " AWS_REGION
AWS_REGION=${AWS_REGION:-eu-central-1}

read -p "Cloudflare API Token: " CLOUDFLARE_API_TOKEN

read -p "Cloudflare Account ID: " CLOUDFLARE_ACCOUNT_ID

read -p "ECR Registry URL (e.g., 123456789012.dkr.ecr.eu-central-1.amazonaws.com): " ECR_REGISTRY

BUILDER_IMAGE="${ECR_REGISTRY}/astro-builder:latest"
UNPUBLISH_IMAGE="${ECR_REGISTRY}/unpublish-image:latest"

# Create environment file
cat > ${ENV_FILE} << ENVEOF
# Redis Configuration
REDIS_ADDR=${REDIS_ADDR}
QUEUE_KEY=${QUEUE_KEY}

# AWS Configuration
AWS_DEFAULT_REGION=${AWS_REGION}

# Cloudflare Configuration
CLOUDFLARE_API_TOKEN=${CLOUDFLARE_API_TOKEN}
CLOUDFLARE_ACCOUNT_ID=${CLOUDFLARE_ACCOUNT_ID}

# Container Images (from ECR)
BUILDER_IMAGE=${BUILDER_IMAGE}
UNPUBLISH_IMAGE=${UNPUBLISH_IMAGE}
ENVEOF

# Set proper permissions
chmod 600 ${ENV_FILE}
chown root:root ${ENV_FILE}

echo ""
echo -e "${GREEN}Environment file created successfully at ${ENV_FILE}${NC}"
echo ""
echo "You can now run the deployment script to install the worker."
