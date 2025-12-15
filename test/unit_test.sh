#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}[INFO]${NC} Running unit tests..."

# Change to project root directory
cd "$(dirname "$0")/.."

# Run all unit tests
if go test ./...; then
    echo ""
    echo -e "${GREEN}[SUCCESS]${NC} All unit tests passed!"
    exit 0
else
    echo ""
    echo -e "${RED}[FAILED]${NC} Some unit tests failed"
    exit 1
fi
