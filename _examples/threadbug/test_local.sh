#!/bin/bash
# 本地测试脚本 - 自动化测试流程

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  本地测试脚本 - Thread Bug 演示${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 获取脚本目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 检查 Go
echo -e "${YELLOW}[1/5] 检查 Go 环境...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go 未安装！请先安装 Go: https://golang.org/dl/${NC}"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓ Go 已安装: $GO_VERSION${NC}"
echo ""

# 检查 Python
echo -e "${YELLOW}[2/5] 检查 Python 环境...${NC}"
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}✗ Python3 未安装！${NC}"
    exit 1
fi
PY_VERSION=$(python3 --version)
echo -e "${GREEN}✓ Python 已安装: $PY_VERSION${NC}"
echo ""

# 测试 1: 运行 Bug 版本
echo -e "${YELLOW}[3/5] 测试 Bug 版本 (demo_bug.go)...${NC}"
echo "----------------------------------------"
if go run demo_bug.go 2>&1 | tee /tmp/bug_output.log | tail -10; then
    ERROR_COUNT=$(grep -c "ERROR" /tmp/bug_output.log || echo "0")
    if [ "$ERROR_COUNT" -gt "0" ]; then
        echo -e "${GREEN}✓ Bug 成功复现！发现 $ERROR_COUNT 个错误${NC}"
    else
        echo -e "${YELLOW}⚠ Bug 未复现（可能随机性导致）${NC}"
        echo "   提示：可以多次运行或增加并发数"
    fi
else
    echo -e "${RED}✗ 运行失败${NC}"
fi
echo ""

# 测试 2: 运行修复版本
echo -e "${YELLOW}[4/5] 测试修复版本 (demo_fixed.go)...${NC}"
echo "----------------------------------------"
if go run demo_fixed.go 2>&1 | tee /tmp/fixed_output.log | tail -10; then
    SUCCESS_COUNT=$(grep -c "successful" /tmp/fixed_output.log || echo "0")
    ERROR_COUNT=$(grep -c "ERROR" /tmp/fixed_output.log || echo "0")
    if [ "$ERROR_COUNT" -eq "0" ] && [ "$SUCCESS_COUNT" -gt "0" ]; then
        echo -e "${GREEN}✓ 修复版本运行成功！${NC}"
    else
        echo -e "${YELLOW}⚠ 修复版本可能有意外错误${NC}"
    fi
else
    echo -e "${RED}✗ 运行失败${NC}"
fi
echo ""

# 检查 gopy（可选）
echo -e "${YELLOW}[5/5] 检查 gopy（可选，用于真实绑定测试）...${NC}"
if command -v gopy &> /dev/null; then
    GOPY_VERSION=$(gopy -h 2>&1 | head -1 || echo "installed")
    echo -e "${GREEN}✓ gopy 已安装${NC}"
    echo ""
    echo -e "${BLUE}提示：可以运行完整测试：${NC}"
    echo "  gopy build -output=./out -vm=python3 github.com/go-python/gopy/_examples/threadbug"
    echo "  cd out && python3 ../test_bug.py"
else
    echo -e "${YELLOW}⚠ gopy 未安装（可选）${NC}"
    echo "   安装: go install github.com/go-python/gopy@latest"
fi
echo ""

# 总结
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  测试完成！${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "查看详细输出："
echo "  Bug 版本: cat /tmp/bug_output.log"
echo "  修复版本: cat /tmp/fixed_output.log"
echo ""
echo "更多信息请查看："
echo "  - README.md - 详细说明"
echo "  - LOCAL_TEST.md - 本地测试指南"
echo "  - SUMMARY.md - 总结文档"
