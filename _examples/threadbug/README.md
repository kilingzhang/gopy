# Thread Switching Bug 演示

## 📋 说明

这个示例演示了 gopy 中 Python 回调的线程切换 bug 及其修复方案。

**重要：这些是纯 Go 代码，模拟了 Python C API 的行为，不需要 Python 环境！**

## 🚀 快速运行

```bash
# 运行 Bug 版本（演示问题）
go run demo_bug.go

# 运行修复版本（演示解决方案）
go run demo_fixed.go
```

**只需要 Go，不需要 Python！**

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

### 关键代码对比

**Bug 版本（错误）：**
```go
func callback() {
    // ❌ 没有 LockOSThread()
    gstate := C.PyGILState_Ensure()  // Thread T1
    
    // 危险：可能切换到 Thread T2
    doWork()
    
    C.PyGILState_Release(gstate)  // CRASH!
}
```

**修复版本（正确）：**
```go
func callback() {
    runtime.LockOSThread()  // ✅ 锁定 goroutine
    defer runtime.UnlockOSThread()
    
    gstate := C.PyGILState_Ensure()  // Thread T1
    defer C.PyGILState_Release(gstate)
    
    // 安全：goroutine 被锁定，不会切换
    doWork()
    
    // Release 在同一线程 - 安全！
}
```

## 📊 预期结果

### demo_bug.go
- 可能显示：`❌ ERROR: PyGILState_Release called from wrong thread!`
- 或者偶尔成功（当 goroutine 没有切换线程时）

### demo_fixed.go
- 应该显示：`✓ Perfect! All callbacks completed successfully!`
- 100% 成功率

## 🔍 代码说明

这两个文件是**纯 Go 代码**，通过模拟 Python C API 的行为来演示问题：

- `SimulatePyGILState_Ensure()` - 模拟 `PyGILState_Ensure()`
- `SimulatePyGILState_Release()` - 模拟 `PyGILState_Release()`
- 通过删除 goroutine 的线程映射来模拟线程切换

**不需要真实的 Python 环境！**

## 💡 关于真实环境

### 这些演示文件（demo_bug.go / demo_fixed.go）
- ✅ **不需要 Python**
- ✅ **不需要 gopy**
- ✅ **只需要 Go**
- ✅ **纯 Go 模拟，演示概念**

### 真实的 gopy 项目
如果你要在真实的 gopy 项目中应用修复：
- 需要 Python 环境
- 需要 gopy 工具
- 需要在生成的代码中添加 `runtime.LockOSThread()`

但**这个演示不需要**，它只是用 Go 代码模拟了问题！

## 📚 文件说明

- **`demo_bug.go`** - Bug 版本演示（没有 `LockOSThread()`）
- **`demo_fixed.go`** - 修复版本演示（使用 `LockOSThread()`）
- **`README.md`** - 本文件

## 🎯 验证

运行后应该看到：
- Bug 版本：可能出现线程切换错误
- 修复版本：所有调用都成功

## ❓ FAQ

**Q: 需要安装 Python 吗？**
A: 不需要！这是纯 Go 代码，模拟了 Python C API 的行为。

**Q: 需要 gopy 吗？**
A: 不需要！这只是概念演示。

**Q: 真实的 bug 需要 Python 环境吗？**
A: 是的，但这里的演示不需要。这个演示用 Go 代码模拟了相同的问题。

**Q: 如何应用到真实项目？**
A: 在 gopy 生成的代码中，找到 `PyGILState_Ensure()` 的位置，在前面添加 `runtime.LockOSThread()`。
