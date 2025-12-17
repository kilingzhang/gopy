#!/bin/bash
# Build and test script for thread switching bug reproduction

set -e

echo "=========================================="
echo "Thread Switching Bug Reproduction Script"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if gopy is installed
if ! command -v gopy &> /dev/null; then
    echo -e "${RED}ERROR: gopy not found. Please install it first:${NC}"
    echo "  go install github.com/go-python/gopy@latest"
    exit 1
fi

# Check if python3 is available
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}ERROR: python3 not found${NC}"
    exit 1
fi

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "Step 1: Building buggy version (without LockOSThread)..."
echo "--------------------------------------------------------"
rm -rf out_bug
gopy build -output=./out_bug -vm=python3 github.com/go-python/gopy/_examples/threadbug

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Build successful${NC}"
echo ""

echo "Step 2: Testing buggy version..."
echo "--------------------------------------------------------"
cd out_bug
echo ""
echo -e "${YELLOW}Running test_bug.py - this may crash!${NC}"
echo ""

# Run the test and capture output
if python3 ../test_bug.py 2>&1 | tee ../test_bug_output.log; then
    echo ""
    echo -e "${GREEN}Test completed (may have worked by chance)${NC}"
else
    echo ""
    echo -e "${RED}Test crashed or failed - bug reproduced!${NC}"
    echo "Check test_bug_output.log for details"
fi

cd ..

echo ""
echo "=========================================="
echo "Next Steps:"
echo "=========================================="
echo ""
echo "To test the FIXED version:"
echo "1. Apply the fix to bind/symbols.go (add runtime.LockOSThread)"
echo "2. Rebuild: gopy build -output=./out_fixed -vm=python3 ..."
echo "3. Run: python3 test_fixed.py"
echo ""
echo "See README.md for detailed instructions"
