# Thread Switching Bug - 真实测试

## 📋 说明

这是**真实的完整流程**测试，使用真实的 gopy 绑定和 Python 环境来复现和验证线程切换 bug。

## 🔧 环境要求

### 必需
- ✅ Go 1.15+
- ✅ Python 3.x
- ✅ gopy 工具
- ✅ pybindgen

### 安装步骤

```bash
# 1. 安装 gopy
go install github.com/go-python/gopy@latest

# 2. 确保 gopy 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin

# 3. 安装 pybindgen
pip3 install pybindgen

# 4. 验证安装
gopy help
python3 -c "import pybindgen"
```

## 🚀 完整测试流程

### 步骤 1: 构建 gopy 绑定

```bash
cd _examples/threadbug

# 方法 1: 使用构建脚本（推荐）
./build.sh

# 方法 2: 手动构建
gopy build -output=./out -vm=python3 \
    github.com/go-python/gopy/_examples/threadbug
```

### 步骤 2: 运行测试

```bash
cd out
python3 ../test_bug.py
```

## 🐛 Bug 说明

### 问题

当 Go goroutine 调用 Python 回调时：
1. `PyGILState_Ensure()` 在 OS 线程 T1 上调用（在 gopy 生成的 C 代码中）
2. Go 调度器可能将 goroutine 切换到 OS 线程 T2
3. `PyGILState_Release()` 在 OS 线程 T2 上调用
4. **崩溃！** Python C API 检测到线程不匹配，导致 SEGFAULT

### 错误信息

可能看到的错误：
```
Fatal Python error: _PyInterpreterState_Get(): no current thread state
Segmentation fault
```

或者：
```
runtime error: invalid memory address or nil pointer dereference
```

## ✅ 修复方案

修复需要在 gopy 生成的代码中添加 `runtime.LockOSThread()`。

### 修复位置

在 `bind/symbols.go` 的 `addSignatureType` 函数中：

```go
// 修复前（buggy）
py2g += "_gstate := C.PyGILState_Ensure()\n"
// ... Python 调用 ...
py2g += "C.PyGILState_Release(_gstate)\n"

// 修复后（fixed）
py2g += "runtime.LockOSThread()\n"
py2g += "defer runtime.UnlockOSThread()\n"
py2g += "_gstate := C.PyGILState_Ensure()\n"
// ... Python 调用 ...
py2g += "C.PyGILState_Release(_gstate)\n"
py2g += "runtime.UnlockOSThread()\n"
```

还需要在生成的 Go 文件中添加：
```go
import "runtime"
```

## 📊 预期结果

### Bug 版本（未修复）
- 可能随机崩溃
- 高并发时更容易崩溃
- 错误信息：segfault 或 thread state 错误

### 修复版本（已修复）
- 应该稳定运行
- 所有回调都成功
- 无崩溃

## 📚 文件说明

- **`callback.go`** - Go 包代码（会被 gopy 绑定）
- **`test_bug.py`** - Python 测试脚本（使用真实绑定）
- **`build.sh`** - 构建脚本
- **`README.md`** - 本文件

## 🎯 测试场景

测试脚本包含三个场景：

1. **基本回调** - 单个调用（通常能工作）
2. **并发回调** - 50 个并发 goroutine（可能崩溃）
3. **高竞争** - 200 个并发 goroutine（更容易崩溃）

## ❓ 常见问题

**Q: 为什么有时不崩溃？**
A: 线程切换是随机的。增加并发数可以提高复现概率。

**Q: 需要特定 Python 版本吗？**
A: Python 3.x 都可以。但构建和运行必须使用同一个 Python 版本。

**Q: 如何确认 bug 已复现？**
A: 看到 segfault 或 "no current thread state" 错误。

**Q: 如何应用修复？**
A: 修改 gopy 源码中的 `bind/symbols.go`，然后重新构建绑定。

## 🔍 调试技巧

### 增加崩溃概率

1. 增加并发数（修改 `test_bug.py` 中的数量）
2. 增加 Python 回调的延迟
3. 在 Go 代码中添加更多 `runtime.Gosched()`

### 查看详细错误

```bash
# 使用 gdb 调试
gdb python3
(gdb) run test_bug.py
(gdb) bt  # 查看堆栈
```

## 📝 注意事项

1. **Python 版本一致性**: 构建和运行必须使用同一个 Python
2. **GIL 状态**: 确保 Python 已正确初始化
3. **线程安全**: 这个 bug 只在多线程场景下出现

## 🎓 学习价值

这个真实测试帮助你理解：
- Go 和 Python 的线程模型差异
- gopy 的工作原理
- CGO 和 Python C API 的交互
- 跨语言调用的陷阱
