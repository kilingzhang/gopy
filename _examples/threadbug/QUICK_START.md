# Quick Start Guide

## 快速复现 Bug

### 方法 1: 运行演示程序（最简单）

```bash
cd _examples/threadbug

# 运行 Bug 版本（会崩溃）
go run demo_bug.go

# 运行修复版本（正常工作）
go run demo_fixed.go

# 或者运行快速对比
./run_demo.sh
```

### 方法 2: 使用 gopy 生成的真实绑定

```bash
cd _examples/threadbug

# 1. 构建 Go 包绑定
gopy build -output=./out -vm=python3 github.com/go-python/gopy/_examples/threadbug

# 2. 运行 Python 测试
cd out
python3 ../test_bug.py
```

## 预期结果

### Bug 版本 (demo_bug.go)
```
❌ ERROR: goroutine X: FATAL ERROR: PyGILState_Release called from wrong thread!
  State was created on thread T1
  Release called on thread T2
  This would cause SEGFAULT in real Python C API!
```

**结果**: 大部分或全部 goroutine 会崩溃

### 修复版本 (demo_fixed.go)
```
✓ Perfect! All callbacks completed successfully!
  runtime.LockOSThread() prevents thread switching
  and ensures thread safety.
```

**结果**: 所有 goroutine 成功完成

## 代码对比

### Bug 版本关键代码
```go
func callback() {
    // ❌ 没有 LockOSThread()
    gstate := PyGILState_Ensure()  // Thread T1
    
    // Goroutine 可能切换到 Thread T2
    runtime.Gosched()
    
    PyGILState_Release(gstate)  // CRASH! 在 Thread T2 上释放
}
```

### 修复版本关键代码
```go
func callback() {
    runtime.LockOSThread()  // ✅ 锁定 goroutine 到当前线程
    defer runtime.UnlockOSThread()
    
    gstate := PyGILState_Ensure()  // Thread T1
    defer PyGILState_Release(gstate)
    
    // Goroutine 被锁定，不会切换线程
    runtime.Gosched()  // 安全！
    
    // Release 在同一线程 T1 上执行 - 安全！
}
```

## 验证清单

- [ ] Bug 版本运行后出现 "wrong thread" 错误
- [ ] 修复版本运行后显示 "All callbacks completed successfully"
- [ ] 理解为什么需要 `runtime.LockOSThread()`
- [ ] 理解 Go goroutine 的线程切换机制

## 进一步学习

- 阅读 `README.md` 了解详细说明
- 查看 `COMPARISON.md` 了解对比结果
- 阅读 Go runtime 文档了解 `LockOSThread()`
