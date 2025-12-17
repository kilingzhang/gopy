# Bug vs Fix Comparison

## Bug Reproduction Results

### Buggy Version (demo_bug.go)
```
❌ ERROR: goroutine 7: FATAL ERROR: PyGILState_Release called from wrong thread!
  State was created on thread 202076
  Release called on thread 596174
  This would cause SEGFAULT in real Python C API!
```

**Result**: 10 out of 10 goroutines crashed! 100% crash rate.

### Fixed Version (demo_fixed.go)
```
✓ Perfect! All callbacks completed successfully!
  runtime.LockOSThread() prevents thread switching
  and ensures thread safety.
```

**Result**: 100 out of 100 goroutines succeeded! 0% crash rate.

## Key Differences

### Buggy Code
```go
func SimulatePythonCallback_BUGGY(msg string) {
    // NO LockOSThread()!
    
    gstate := SimulatePyGILState_Ensure()  // Thread T1
    
    // Goroutine may switch to Thread T2 here!
    runtime.Gosched()
    time.Sleep(time.Millisecond)
    
    SimulatePyGILState_Release(gstate)  // CRASH! Wrong thread!
}
```

### Fixed Code
```go
func SimulatePythonCallback_FIXED(msg string) {
    runtime.LockOSThread()  // ✅ Lock goroutine to thread
    defer runtime.UnlockOSThread()
    
    gstate := SimulatePyGILState_Ensure()  // Thread T1
    defer SimulatePyGILState_Release(gstate)
    
    // Goroutine CANNOT switch threads now!
    runtime.Gosched()  // Safe - stays on T1
    time.Sleep(time.Millisecond)
    
    // Release happens on same thread T1 - SAFE!
}
```

## How to Run

### Reproduce the Bug
```bash
cd _examples/threadbug
go run demo_bug.go
```

Expected: Multiple crashes with "wrong thread" errors

### Test the Fix
```bash
cd _examples/threadbug
go run demo_fixed.go
```

Expected: All callbacks complete successfully

### Quick Demo
```bash
cd _examples/threadbug
./run_demo.sh
```

## Real-World Impact

In a real gopy application:
- **Without fix**: Web servers crash randomly under load
- **With fix**: Stable operation even under high concurrency

The fix is critical for production use!
