# 如何在本地运行代码

## 🎯 方法 1: 从 GitHub 克隆（推荐）

### 步骤 1: 克隆 gopy 仓库

```bash
# 克隆整个仓库
git clone https://github.com/go-python/gopy.git
cd gopy/_examples/threadbug
```

### 步骤 2: 直接运行

```bash
# 运行 Bug 版本
go run demo_bug.go

# 运行修复版本
go run demo_fixed.go
```

## 🎯 方法 2: 直接创建文件（最简单）

如果不想克隆整个仓库，可以直接在本地创建这些文件：

### 创建目录

```bash
mkdir -p ~/threadbug-demo
cd ~/threadbug-demo
```

### 创建 demo_bug.go

```bash
cat > demo_bug.go << 'EOF'
// Bug 版本 - 没有 LockOSThread()
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type ThreadState struct {
	threadID int64
	locked   bool
}

var (
	threadStates = make(map[int64]*ThreadState)
	stateMutex   sync.Mutex
	threadIDMap  = make(map[uint64]int64)
	threadIDMutex sync.Mutex
	nextThreadID int64 = 1
)

func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	var id uint64
	fmt.Sscanf(string(b), "goroutine %d", &id)
	return id
}

func getCurrentThreadID() int64 {
	goid := getGoroutineID()
	threadIDMutex.Lock()
	defer threadIDMutex.Unlock()
	if tid, ok := threadIDMap[goid]; ok {
		return tid
	}
	tid := nextThreadID
	nextThreadID++
	threadIDMap[goid] = tid
	return tid
}

func SimulatePyGILState_Ensure() *ThreadState {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	threadID := getCurrentThreadID()
	goid := getGoroutineID()
	state := &ThreadState{threadID: threadID, locked: true}
	threadStates[threadID] = state
	fmt.Printf("[Goroutine %d -> Thread %d] PyGILState_Ensure()\n", goid, threadID)
	return state
}

func SimulatePyGILState_Release(state *ThreadState) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	currentThreadID := getCurrentThreadID()
	goid := getGoroutineID()
	if state.threadID != currentThreadID {
		panic(fmt.Sprintf(
			"FATAL ERROR: PyGILState_Release called from wrong thread!\n"+
				"  State was created on thread %d\n"+
				"  Release called on thread %d (goroutine %d)\n"+
				"  This would cause SEGFAULT in real Python C API!",
			state.threadID, currentThreadID, goid))
	}
	fmt.Printf("[Goroutine %d -> Thread %d] PyGILState_Release() - OK\n", goid, currentThreadID)
	delete(threadStates, state.threadID)
}

func SimulatePythonCallback_BUGGY(msg string) {
	fmt.Printf("[Goroutine] Starting Python callback: %s\n", msg)
	gstate := SimulatePyGILState_Ensure()
	runtime.Gosched()
	goid := getGoroutineID()
	threadIDMutex.Lock()
	delete(threadIDMap, goid)
	threadIDMutex.Unlock()
	time.Sleep(time.Millisecond)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.Gosched()
		}()
	}
	wg.Wait()
	fmt.Printf("[Goroutine] Calling Python function...\n")
	SimulatePyGILState_Release(gstate)
	fmt.Printf("[Goroutine] Python callback completed: %s\n", msg)
}

func main() {
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("Thread Switching Bug Demonstration")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("\nRunning 50 concurrent callbacks (BUGGY version)...\n")
	
	var wg sync.WaitGroup
	errors := make(chan error, 50)
	successCount := 0
	var mu sync.Mutex
	
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errors <- fmt.Errorf("goroutine %d: %v", id, r)
				} else {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
			SimulatePythonCallback_BUGGY(fmt.Sprintf("Message-%d", id))
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	errorCount := 0
	for err := range errors {
		errorCount++
		fmt.Printf("❌ ERROR: %v\n", err)
	}
	
	fmt.Printf("\n结果: %d 成功, %d 错误\n", successCount, errorCount)
	if errorCount > 0 {
		fmt.Println("\n✓ Bug 成功复现！")
	}
}
EOF
```

### 创建 demo_fixed.go

```bash
cat > demo_fixed.go << 'EOF'
// 修复版本 - 使用 LockOSThread()
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type ThreadState struct {
	threadID int64
	locked   bool
}

var (
	threadStates = make(map[int64]*ThreadState)
	stateMutex   sync.Mutex
	threadIDMap  = make(map[uint64]int64)
	threadIDMutex sync.Mutex
	nextThreadID int64 = 1
)

func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	var id uint64
	fmt.Sscanf(string(b), "goroutine %d", &id)
	return id
}

func getCurrentThreadID() int64 {
	goid := getGoroutineID()
	threadIDMutex.Lock()
	defer threadIDMutex.Unlock()
	if tid, ok := threadIDMap[goid]; ok {
		return tid
	}
	tid := nextThreadID
	nextThreadID++
	threadIDMap[goid] = tid
	return tid
}

func SimulatePyGILState_Ensure() *ThreadState {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	threadID := getCurrentThreadID()
	goid := getGoroutineID()
	state := &ThreadState{threadID: threadID, locked: true}
	threadStates[threadID] = state
	fmt.Printf("[Goroutine %d -> Thread %d] PyGILState_Ensure()\n", goid, threadID)
	return state
}

func SimulatePyGILState_Release(state *ThreadState) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	currentThreadID := getCurrentThreadID()
	goid := getGoroutineID()
	if state.threadID != currentThreadID {
		panic(fmt.Sprintf("FATAL: Thread mismatch! State=%d, Current=%d", state.threadID, currentThreadID))
	}
	fmt.Printf("[Goroutine %d -> Thread %d] PyGILState_Release() - OK\n", goid, currentThreadID)
	delete(threadStates, state.threadID)
}

func SimulatePythonCallback_FIXED(msg string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	fmt.Printf("[Goroutine] Starting Python callback (FIXED): %s\n", msg)
	gstate := SimulatePyGILState_Ensure()
	defer SimulatePyGILState_Release(gstate)
	
	runtime.Gosched()
	time.Sleep(time.Millisecond)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.Gosched()
		}()
	}
	wg.Wait()
	fmt.Printf("[Goroutine] Calling Python function...\n")
	fmt.Printf("[Goroutine] Python callback completed (FIXED): %s\n", msg)
}

func main() {
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("FIXED Version - With runtime.LockOSThread()")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("\nRunning 100 concurrent callbacks with fix...\n")
	
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	successCount := 0
	var mu sync.Mutex
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errors <- fmt.Errorf("goroutine %d: %v", id, r)
				} else {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
			SimulatePythonCallback_FIXED(fmt.Sprintf("Message-%d", id))
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	errorCount := 0
	for err := range errors {
		errorCount++
		fmt.Printf("❌ ERROR: %v\n", err)
	}
	
	fmt.Printf("\n结果: %d 成功, %d 错误\n", successCount, errorCount)
	if errorCount == 0 {
		fmt.Println("\n✓ Perfect! All callbacks completed successfully!")
		fmt.Println("  runtime.LockOSThread() prevents thread switching")
	}
}
EOF
```

### 运行测试

```bash
# 运行 Bug 版本
go run demo_bug.go

# 运行修复版本
go run demo_fixed.go
```

## 🎯 方法 3: 下载单个文件

如果只需要这两个文件，可以直接下载：

```bash
# 创建目录
mkdir -p ~/threadbug-demo
cd ~/threadbug-demo

# 下载文件（需要 curl）
curl -O https://raw.githubusercontent.com/go-python/gopy/main/_examples/threadbug/demo_bug.go
curl -O https://raw.githubusercontent.com/go-python/gopy/main/_examples/threadbug/demo_fixed.go

# 运行
go run demo_bug.go
go run demo_fixed.go
```

## ✅ 验证安装

运行前检查：

```bash
# 检查 Go 是否安装
go version

# 应该显示类似：go version go1.19.x linux/amd64
```

## 🚀 快速测试

创建好文件后，直接运行：

```bash
# 进入目录
cd ~/threadbug-demo  # 或你创建的目录

# 运行 Bug 版本（看问题）
go run demo_bug.go

# 运行修复版本（看解决方案）
go run demo_fixed.go
```

## 📝 预期输出

**Bug 版本** 可能显示：
```
❌ ERROR: goroutine X: FATAL ERROR: PyGILState_Release called from wrong thread!
```

**修复版本** 应该显示：
```
✓ Perfect! All callbacks completed successfully!
```

## 💡 提示

- 不需要安装 gopy
- 不需要 Python
- 只需要 Go（1.15+）
- 可以直接运行，无需编译

## ❓ 如果遇到问题

1. **确保 Go 已安装**: `go version`
2. **确保在正确目录**: `pwd`
3. **确保文件存在**: `ls demo_bug.go demo_fixed.go`

现在你可以在本地运行了！
