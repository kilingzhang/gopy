# 线程切换 Bug 演示 - 中文说明

## 🎯 快速开始

### 最简单的测试方法（推荐）

```bash
# 1. 进入目录
cd _examples/threadbug

# 2. 运行 Bug 演示
go run demo_bug.go

# 3. 运行修复演示
go run demo_fixed.go
```

就这么简单！不需要安装 gopy 或其他依赖，只需要 Go。

## 📁 文件说明

### 核心文件

- **`demo_bug.go`** - Bug 版本演示（没有 `runtime.LockOSThread()`）
- **`demo_fixed.go`** - 修复版本演示（使用 `runtime.LockOSThread()`）

这两个文件可以直接运行，模拟了真实的 bug 场景。

### 其他文件

- **`callback.go`** - Go 包代码（用于 gopy 生成 Python 绑定）
- **`test_bug.py`** / **`test_fixed.py`** - Python 测试脚本
- **`LOCAL_TEST.md`** - 详细的本地测试指南
- **`SUMMARY.md`** - 问题总结

## 🐛 Bug 说明

### 问题

当 Go goroutine 调用 Python 回调时：
1. `PyGILState_Ensure()` 在 OS 线程 T1 上调用
2. Go 调度器可能将 goroutine 切换到 OS 线程 T2  
3. `PyGILState_Release()` 在 OS 线程 T2 上调用
4. **崩溃！** Python C API 检测到线程不匹配

### 原因

- Go 使用 M:N 调度（goroutine 可以在线程间切换）
- Python GIL 绑定到特定 OS 线程
- 两者不兼容！

## ✅ 修复方案

### 关键代码

**Bug 版本（错误）：**
```go
func callback() {
    // ❌ 没有锁定
    gstate := C.PyGILState_Ensure()  // 线程 T1
    
    // 危险：可能切换到线程 T2
    doWork()
    
    C.PyGILState_Release(gstate)  // CRASH!
}
```

**修复版本（正确）：**
```go
func callback() {
    runtime.LockOSThread()  // ✅ 锁定 goroutine
    defer runtime.UnlockOSThread()
    
    gstate := C.PyGILState_Ensure()  // 线程 T1
    defer C.PyGILState_Release(gstate)
    
    // 安全：goroutine 被锁定，不会切换
    doWork()
    
    // Release 在同一线程 - 安全！
}
```

## 🧪 测试步骤

### 方法 1: 直接运行（最简单）

```bash
cd _examples/threadbug
go run demo_bug.go      # 看 bug
go run demo_fixed.go    # 看修复
```

### 方法 2: 使用自动化脚本

```bash
cd _examples/threadbug
chmod +x test_local.sh
./test_local.sh
```

### 方法 3: 完整测试（需要 gopy）

```bash
# 1. 安装 gopy
go install github.com/go-python/gopy@latest

# 2. 构建绑定
cd /path/to/gopy
gopy build -output=./_examples/threadbug/out -vm=python3 \
    ./_examples/threadbug

# 3. 运行 Python 测试
cd _examples/threadbug/out
python3 ../test_bug.py
```

## 📊 预期结果

### Bug 版本
- 可能显示：`❌ ERROR: PyGILState_Release called from wrong thread!`
- 或者偶尔成功（当 goroutine 没有切换线程时）

### 修复版本
- 应该显示：`✓ Perfect! All callbacks completed successfully!`
- 100% 成功率

## 🔍 理解要点

1. **Go Runtime**: Goroutine 可以在不同 OS 线程间切换
2. **Python GIL**: 绑定到特定 OS 线程，不能跨线程
3. **解决方案**: `runtime.LockOSThread()` 锁定 goroutine 到当前线程

## 📚 更多信息

- **`LOCAL_TEST.md`** - 详细的本地测试指南（包含故障排除）
- **`SUMMARY.md`** - 问题总结和代码对比
- **`QUICK_START.md`** - 快速开始指南

## ❓ 常见问题

### Q: 为什么 Bug 版本有时不崩溃？
A: 因为线程切换是随机的。增加并发数可以提高复现概率。

### Q: 需要安装 gopy 吗？
A: 不需要！`demo_bug.go` 和 `demo_fixed.go` 可以直接运行。

### Q: 如何应用到真实项目？
A: 在 gopy 生成的代码中，找到 `PyGILState_Ensure()` 的位置，在前面添加 `runtime.LockOSThread()`。

## 🎓 学习价值

这个演示帮助你理解：
- Go 的并发模型
- Python C API 的线程安全
- 跨语言调用的陷阱
- `runtime.LockOSThread()` 的作用
