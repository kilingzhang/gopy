# 本地测试指南

## 📋 前置要求

### 1. 安装 Go
```bash
# 检查 Go 版本（需要 1.15+）
go version

# 如果没有安装，访问：https://golang.org/dl/
```

### 2. 安装 Python 3
```bash
# 检查 Python 版本
python3 --version

# 如果没有安装：
# macOS: brew install python3
# Ubuntu: sudo apt-get install python3
# Windows: 从 python.org 下载安装
```

### 3. 安装 gopy（如果需要测试真实绑定）
```bash
# 安装 gopy
go install github.com/go-python/gopy@latest

# 确保 $GOPATH/bin 或 $HOME/go/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin

# 验证安装
gopy help
```

### 4. 安装 pybindgen（如果需要测试真实绑定）
```bash
# 安装 pybindgen
pip3 install pybindgen

# 验证安装
python3 -c "import pybindgen; print('OK')"
```

## 🚀 快速测试（推荐）

### 方法 1: 运行 Go 演示程序（最简单，无需 gopy）

这是最简单的方法，直接运行 Go 程序模拟 bug：

```bash
# 1. 进入目录
cd /workspace/_examples/threadbug

# 2. 运行 Bug 版本（可能崩溃）
go run demo_bug.go

# 3. 运行修复版本（应该成功）
go run demo_fixed.go

# 4. 或者运行快速对比脚本
chmod +x run_demo.sh
./run_demo.sh
```

**预期输出：**

Bug 版本可能显示：
```
❌ ERROR: goroutine X: FATAL ERROR: PyGILState_Release called from wrong thread!
```

修复版本应该显示：
```
✓ Perfect! All callbacks completed successfully!
```

## 🔧 完整测试（使用真实 gopy 绑定）

如果你想测试真实的 Python 绑定，需要以下步骤：

### 步骤 1: 准备 Go 模块

```bash
cd /workspace/_examples/threadbug

# 初始化 Go 模块（如果还没有）
go mod init github.com/go-python/gopy/_examples/threadbug

# 或者如果已经在 gopy 项目中，确保在项目根目录
cd /workspace
```

### 步骤 2: 构建 Python 绑定

```bash
cd /workspace/_examples/threadbug

# 构建绑定（输出到 out 目录）
gopy build -output=./out -vm=python3 \
    github.com/go-python/gopy/_examples/threadbug

# 如果遇到模块路径问题，可以这样：
cd /workspace
gopy build -output=./_examples/threadbug/out -vm=python3 \
    ./_examples/threadbug
```

### 步骤 3: 运行 Python 测试

```bash
cd /workspace/_examples/threadbug/out

# 运行 Bug 测试
python3 ../test_bug.py

# 运行修复测试（需要先应用修复）
python3 ../test_fixed.py
```

## 🐛 常见问题解决

### 问题 1: `go: cannot find module`

**解决方案：**
```bash
# 确保在正确的目录
cd /workspace

# 或者初始化模块
cd /workspace/_examples/threadbug
go mod init test-threadbug
go mod tidy
```

### 问题 2: `gopy: command not found`

**解决方案：**
```bash
# 安装 gopy
go install github.com/go-python/gopy@latest

# 检查 PATH
echo $PATH | grep -q "$(go env GOPATH)/bin" || export PATH=$PATH:$(go env GOPATH)/bin

# 验证
which gopy
```

### 问题 3: `pybindgen not found`

**解决方案：**
```bash
# 安装 pybindgen
pip3 install pybindgen

# 验证
python3 -c "import pybindgen"
```

### 问题 4: Python 版本不匹配

**解决方案：**
```bash
# 明确指定 Python 版本
gopy build -output=./out -vm=/usr/bin/python3.9 ...

# 或者
gopy build -output=./out -vm=python3.10 ...
```

### 问题 5: 权限问题（Linux）

**解决方案：**
```bash
# 给脚本添加执行权限
chmod +x run_demo.sh
chmod +x build_and_test.sh

# 如果遇到库加载问题
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:.
```

## 📝 测试步骤总结

### 最简单的测试（5分钟）

```bash
# 1. 克隆或下载代码
cd /path/to/gopy/_examples/threadbug

# 2. 运行演示
go run demo_bug.go      # 看 bug
go run demo_fixed.go    # 看修复
```

### 完整测试（15分钟）

```bash
# 1. 安装依赖
go install github.com/go-python/gopy@latest
pip3 install pybindgen

# 2. 构建绑定
cd /path/to/gopy
gopy build -output=./_examples/threadbug/out -vm=python3 \
    ./_examples/threadbug

# 3. 运行测试
cd _examples/threadbug/out
python3 ../test_bug.py
```

## 🎯 验证清单

运行测试后，确认：

- [ ] `demo_bug.go` 运行后可能出现错误
- [ ] `demo_fixed.go` 运行后显示成功
- [ ] 理解代码中的关键差异
- [ ] 理解为什么需要 `runtime.LockOSThread()`

## 💡 调试技巧

### 增加线程切换概率

如果 bug 版本没有崩溃，可以：

1. **增加并发数**：修改 `demo_bug.go` 中的 goroutine 数量
2. **增加延迟**：在 `doWork()` 中添加 `time.Sleep()`
3. **强制切换**：多次调用 `runtime.Gosched()`

### 查看详细输出

```bash
# 查看完整输出
go run demo_bug.go 2>&1 | tee output.log

# 只查看错误
go run demo_bug.go 2>&1 | grep ERROR

# 统计错误数量
go run demo_bug.go 2>&1 | grep -c ERROR
```

## 📚 下一步

- 阅读 `README.md` 了解详细原理
- 查看 `SUMMARY.md` 了解总结
- 阅读 `COMPARISON.md` 了解对比

## 🆘 需要帮助？

如果遇到问题：
1. 检查 Go 和 Python 版本
2. 确认所有依赖已安装
3. 查看错误信息
4. 参考常见问题部分
