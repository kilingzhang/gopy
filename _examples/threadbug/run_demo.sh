#!/bin/bash
# Quick demo script to show the bug and fix

set -e

echo "=========================================="
echo "Thread Switching Bug - Quick Demo"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo -e "${YELLOW}Demo 1: BUGGY Version (may crash)${NC}"
echo "----------------------------------------"
echo "This simulates the bug where goroutines switch threads"
echo "between PyGILState_Ensure() and PyGILState_Release()"
echo ""
go run demo_bug.go 2>&1 | head -100

echo ""
echo ""
echo -e "${YELLOW}Demo 2: FIXED Version (should work)${NC}"
echo "----------------------------------------"
echo "This uses runtime.LockOSThread() to prevent thread switching"
echo ""
go run demo_fixed.go 2>&1 | tail -20

echo ""
echo -e "${GREEN}Demo completed!${NC}"
echo ""
echo "The buggy version may crash randomly when goroutines switch threads."
echo "The fixed version always works because goroutines are locked to their threads."
