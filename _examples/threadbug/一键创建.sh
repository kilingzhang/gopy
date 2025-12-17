#!/bin/bash
# 一键创建本地测试文件

echo "=========================================="
echo "创建本地测试文件"
echo "=========================================="
echo ""

# 创建目录
DIR="$HOME/threadbug-demo"
mkdir -p "$DIR"
cd "$DIR"

echo "创建目录: $DIR"
echo ""

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装！"
    echo "请访问: https://golang.org/dl/"
    exit 1
fi

echo "✓ Go 已安装: $(go version)"
echo ""

# 提示：由于文件较大，这里提供下载链接
echo "由于文件较大，请选择以下方式之一："
echo ""
echo "方法 1: 从 GitHub 下载"
echo "  curl -O https://raw.githubusercontent.com/go-python/gopy/main/_examples/threadbug/demo_bug.go"
echo "  curl -O https://raw.githubusercontent.com/go-python/gopy/main/_examples/threadbug/demo_fixed.go"
echo ""
echo "方法 2: 克隆整个仓库"
echo "  git clone https://github.com/go-python/gopy.git"
echo "  cd gopy/_examples/threadbug"
echo ""
echo "方法 3: 查看 GET_STARTED.md 获取完整文件内容"
echo ""
echo "创建完成后，运行："
echo "  go run demo_bug.go"
echo "  go run demo_fixed.go"
