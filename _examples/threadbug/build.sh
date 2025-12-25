#!/bin/bash
# 构建脚本 - 生成 gopy 绑定

set -e

echo "=========================================="
echo "构建 gopy 绑定"
echo "=========================================="
echo ""

# 检查 gopy
if ! command -v gopy &> /dev/null; then
    echo "ERROR: gopy not found!"
    echo "Install with: go install github.com/go-python/gopy@latest"
    exit 1
fi

# 检查 Python
if ! command -v python3 &> /dev/null; then
    echo "ERROR: python3 not found!"
    exit 1
fi

# 获取脚本目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 确定输出目录
OUTPUT_DIR="./out"

echo "Building bindings..."
echo "  Package: github.com/go-python/gopy/_examples/threadbug"
echo "  Output:  $OUTPUT_DIR"
echo "  Python:  $(python3 --version)"
echo ""

# 清理旧输出
rm -rf "$OUTPUT_DIR"

# 构建绑定
gopy build -output="$OUTPUT_DIR" -vm=python3 \
    github.com/go-python/gopy/_examples/threadbug

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Build successful!"
    echo ""
    echo "Next steps:"
    echo "  cd $OUTPUT_DIR"
    echo "  python3 ../test_bug.py"
else
    echo ""
    echo "✗ Build failed!"
    exit 1
fi
