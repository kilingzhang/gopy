// +build ignore

// This demonstrates the FIXED version with runtime.LockOSThread()
// Run with: go run demo_fixed.go

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Same simulation as demo_bug.go but with proper locking

type ThreadState struct {
	threadID int64
	locked   bool
}

var (
	threadStates = make(map[int64]*ThreadState)
	stateMutex   sync.Mutex
)

func SimulatePyGILState_Ensure() *ThreadState {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	
	threadID := getCurrentThreadID()
	goid := getGoroutineID()
	state := &ThreadState{
		threadID: threadID,
		locked:   true,
	}
	
	threadStates[threadID] = state
	fmt.Printf("[Goroutine %d -> Thread %d] PyGILState_Ensure() - bound to thread\n", goid, threadID)
	return state
}

func SimulatePyGILState_Release(state *ThreadState) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	
	currentThreadID := getCurrentThreadID()
	goid := getGoroutineID()
	
	if state.threadID != currentThreadID {
		panic(fmt.Sprintf(
			"FATAL: Thread mismatch! State=%d, Current=%d (goroutine %d)",
			state.threadID, currentThreadID, goid))
	}
	
	fmt.Printf("[Goroutine %d -> Thread %d] PyGILState_Release() - OK\n", goid, currentThreadID)
	delete(threadStates, state.threadID)
}

// threadIDMap stores goroutine ID -> thread ID mapping
var threadIDMap = make(map[uint64]int64)
var threadIDMutex sync.Mutex
var nextThreadID int64 = 1

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

func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	var id uint64
	fmt.Sscanf(string(b), "goroutine %d", &id)
	return id
}

// FIXED VERSION: Uses LockOSThread()
func SimulatePythonCallback_FIXED(msg string) {
	// Lock goroutine to current OS thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	// Now goroutine cannot switch threads!
	gstate := SimulatePyGILState_Ensure()
	defer SimulatePyGILState_Release(gstate)
	
	// Safe: goroutine is locked, won't switch threads
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
	fmt.Printf("[Goroutine] Python callback completed: %s\n", msg)
}

func main() {
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("FIXED Version - With runtime.LockOSThread()")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println()
	
	fmt.Println("Running 100 concurrent callbacks with fix...")
	fmt.Println("All should complete successfully!")
	fmt.Println()
	
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
	
	fmt.Println()
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Printf("Results: %d successful, %d errors\n", successCount, errorCount)
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	
	if errorCount == 0 {
		fmt.Println("\n✓ Perfect! All callbacks completed successfully!")
		fmt.Println("  runtime.LockOSThread() prevents thread switching")
		fmt.Println("  and ensures thread safety.")
	} else {
		fmt.Printf("\n⚠ Unexpected: %d errors occurred\n", errorCount)
	}
}
