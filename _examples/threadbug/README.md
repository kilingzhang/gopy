# Thread Switching Bug Reproduction

This example demonstrates the thread switching bug in gopy when calling Python callbacks from Go goroutines.

## The Bug

When a Go goroutine calls Python callbacks:
1. `PyGILState_Ensure()` is called on OS thread T1
2. Go scheduler may switch the goroutine to OS thread T2
3. `PyGILState_Release()` is called on thread T2
4. **CRASH!** Thread mismatch causes segfault

## The Fix

Use `runtime.LockOSThread()` to lock the goroutine to the current OS thread:
1. Call `runtime.LockOSThread()` before `PyGILState_Ensure()`
2. Goroutine is locked to current thread, cannot switch
3. `PyGILState_Release()` is called on same thread
4. **SUCCESS!** No crash

## Building

### Step 1: Build the Go package bindings (BUGGY version)

```bash
cd _examples/threadbug
gopy build -output=./out_bug -vm=python3 github.com/go-python/gopy/_examples/threadbug
```

### Step 2: Build the Go package bindings (FIXED version)

You need to manually modify the generated code to add `runtime.LockOSThread()` calls, or use the fixed generator.

For now, you can test the buggy version:

```bash
# Test buggy version
cd out_bug
python3 ../test_bug.py
```

## Running Tests

### Test 1: Reproduce the Bug

```bash
cd out_bug
python3 ../test_bug.py
```

This will likely crash or show errors when goroutines switch threads.

### Test 2: Test the Fix

After applying the fix (adding `runtime.LockOSThread()`), run:

```bash
cd out_fixed
python3 ../test_fixed.py
```

This should run without crashes.

## Expected Behavior

### Buggy Version:
- May crash with segfault
- May show "Fatal Python error: _PyInterpreterState_Get(): no current thread state"
- May work sometimes (when goroutines don't switch threads)

### Fixed Version:
- Should always work
- No crashes
- All callbacks execute successfully

## Manual Fix Instructions

To manually apply the fix to the generated code:

1. Find `bind/symbols.go` function `addSignatureType`
2. Add `runtime.LockOSThread()` before `PyGILState_Ensure()`
3. Add `runtime.UnlockOSThread()` after `PyGILState_Release()`
4. Add `import "runtime"` to the generated Go file

Example fix in generated code:

```go
import "runtime"

func pyCallback(_fun_arg *C.PyObject, ...) {
    runtime.LockOSThread()  // ADD THIS
    defer runtime.UnlockOSThread()  // ADD THIS
    
    _gstate := C.PyGILState_Ensure()
    defer C.PyGILState_Release(_gstate)
    
    // ... Python call ...
    
    C.PyGILState_Release(_gstate)
    runtime.UnlockOSThread()  // ADD THIS
}
```
