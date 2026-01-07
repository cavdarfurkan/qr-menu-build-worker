#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="qr-menu-build-worker"
INSTALL_PATH="/usr/local/bin/${BINARY_NAME}"
SERVICE_NAME="${BINARY_NAME}.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"
ENV_FILE="/etc/qr-menu-worker.env"
S3_BUCKET="qr-menu-artifacts"
S3_KEY="build-worker/qr-menu-build-worker-latest"

echo -e "${GREEN}=== QR Menu Build Worker Deployment Script ===${NC}"
echo ""

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
   exit 1
fi

# Step 1: Stop existing service if running
echo -e "${YELLOW}[1/7] Stopping existing service (if running)...${NC}"
if systemctl is-active --quiet ${SERVICE_NAME}; then
    systemctl stop ${SERVICE_NAME}
    echo "Service stopped"
else
    echo "Service not running"
fi

# Step 2: Download binary from S3
echo -e "${YELLOW}[2/7] Downloading binary from S3...${NC}"
aws s3 cp s3://${S3_BUCKET}/${S3_KEY} /tmp/${BINARY_NAME}
chmod +x /tmp/${BINARY_NAME}
echo "Binary downloaded"

# Step 3: Backup old binary if exists
if [ -f "${INSTALL_PATH}" ]; then
    echo -e "${YELLOW}[3/7] Backing up old binary...${NC}"
    cp ${INSTALL_PATH} ${INSTALL_PATH}.backup.$(date +%Y%m%d-%H%M%S)
    echo "Backup created"
else
    echo -e "${YELLOW}[3/7] No existing binary to backup${NC}"
fi

# Step 4: Install new binary
echo -e "${YELLOW}[4/7] Installing new binary...${NC}"
mv /tmp/${BINARY_NAME} ${INSTALL_PATH}
chown root:root ${INSTALL_PATH}
echo "Binary installed at ${INSTALL_PATH}"

# Step 5: Install systemd service file
echo -e "${YELLOW}[5/7] Installing systemd service file...${NC}"
if [ -f "./qr-menu-build-worker.service" ]; then
    cp ./qr-menu-build-worker.service ${SERVICE_PATH}
    echo "Service file installed"
else
    echo -e "${RED}Error: Service file not found in current directory${NC}"
    exit 1
fi

# Step 6: Check environment file
echo -e "${YELLOW}[6/7] Checking environment file...${NC}"
if [ ! -f "${ENV_FILE}" ]; then
    echo -e "${RED}Warning: Environment file ${ENV_FILE} not found!${NC}"
    echo "Please create it before starting the service."
    echo "Example file is available at: ./qr-menu-worker.env.example"
    echo ""
    read -p "Do you want to continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "Environment file exists"
    chmod 600 ${ENV_FILE}
fi

# Step 7: Reload systemd and enable service
echo -e "${YELLOW}[7/7] Reloading systemd and enabling service...${NC}"
systemctl daemon-reload
systemctl enable ${SERVICE_NAME}
echo "Service enabled"

echo ""
echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
echo "Next steps:"
echo "  1. Verify environment file: sudo cat ${ENV_FILE}"
echo "  2. Start the service: sudo systemctl start ${SERVICE_NAME}"
echo "  3. Check status: sudo systemctl status ${SERVICE_NAME}"
echo "  4. View logs: sudo journalctl -u ${SERVICE_NAME} -f"
echo ""
echo "To start the service now, run:"
echo "  sudo systemctl start ${SERVICE_NAME}"
