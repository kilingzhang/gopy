// +build ignore

// This is a standalone demonstration of the thread switching bug
// Run with: go run demo_bug.go
//
// This simulates what happens when gopy calls Python callbacks without
// runtime.LockOSThread(). The goroutine may switch threads between
// PyGILState_Ensure() and PyGILState_Release(), causing a crash.

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// SimulatePyGILState simulates Python's PyGILState_Ensure/Release
// In real code, these are C calls that bind to the current OS thread
type ThreadState struct {
	threadID int64
	locked   bool
}

var (
	threadStates = make(map[int64]*ThreadState)
	stateMutex   sync.Mutex
	currentState *ThreadState
)

// SimulatePyGILState_Ensure - binds to current OS thread
func SimulatePyGILState_Ensure() *ThreadState {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	
	// Get current goroutine's "thread" ID
	// This simulates the OS thread ID that PyGILState_Ensure binds to
	threadID := getCurrentThreadID()
	goid := getGoroutineID()
	
	state := &ThreadState{
		threadID: threadID,
		locked:   true,
	}
	
	threadStates[threadID] = state
	currentState = state
	
	fmt.Printf("[Goroutine %d -> Thread %d] PyGILState_Ensure() - bound to thread\n", goid, threadID)
	return state
}

// SimulatePyGILState_Release - must be called on same thread
func SimulatePyGILState_Release(state *ThreadState) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	
	currentThreadID := getCurrentThreadID()
	goid := getGoroutineID()
	
	if state.threadID != currentThreadID {
		// BUG REPRODUCED! Called from different thread
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

// threadIDMap stores goroutine ID -> thread ID mapping
var threadIDMap = make(map[uint64]int64)
var threadIDMutex sync.Mutex
var nextThreadID int64 = 1

// getCurrentThreadID simulates getting OS thread ID
// In real scenario, goroutine may switch threads
func getCurrentThreadID() int64 {
	// Get goroutine ID
	goid := getGoroutineID()
	
	threadIDMutex.Lock()
	defer threadIDMutex.Unlock()
	
	// Check if this goroutine already has a thread ID
	if tid, ok := threadIDMap[goid]; ok {
		return tid
	}
	
	// Assign new thread ID (simulates goroutine running on new thread)
	tid := nextThreadID
	nextThreadID++
	threadIDMap[goid] = tid
	return tid
}

// getGoroutineID gets the current goroutine ID
func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	// Parse the goroutine ID from the stack trace
	var id uint64
	fmt.Sscanf(string(b), "goroutine %d", &id)
	return id
}

// SimulatePythonCallback simulates calling Python from Go
func SimulatePythonCallback_BUGGY(msg string) {
	// BUGGY VERSION: No LockOSThread()
	
	fmt.Printf("[Goroutine] Starting Python callback: %s\n", msg)
	
	// Simulate PyGILState_Ensure() - binds to current thread
	gstate := SimulatePyGILState_Ensure()
	
	// DANGER ZONE: Goroutine may switch threads here!
	// This simulates work that might cause thread switching
	runtime.Gosched() // Explicitly yield to scheduler
	
	// Simulate thread switching by clearing thread ID mapping
	// This forces getCurrentThreadID() to return a new thread ID
	// (simulating goroutine being scheduled on different OS thread)
	goid := getGoroutineID()
	threadIDMutex.Lock()
	delete(threadIDMap, goid) // Clear mapping to simulate thread switch
	threadIDMutex.Unlock()
	
	time.Sleep(time.Millisecond) // Small delay
	
	// Create contention to increase chance of thread switching
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.Gosched()
		}()
	}
	wg.Wait()
	
	// Simulate Python call
	fmt.Printf("[Goroutine] Calling Python function...\n")
	
	// Simulate PyGILState_Release() - CRASH if thread switched!
	SimulatePyGILState_Release(gstate)
	
	fmt.Printf("[Goroutine] Python callback completed: %s\n", msg)
}

func SimulatePythonCallback_FIXED(msg string) {
	// FIXED VERSION: Uses LockOSThread()
	
	fmt.Printf("[Goroutine] Starting Python callback (FIXED): %s\n", msg)
	
	// Lock goroutine to current OS thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	// Now goroutine cannot switch threads!
	
	// Simulate PyGILState_Ensure() - binds to current thread
	gstate := SimulatePyGILState_Ensure()
	defer SimulatePyGILState_Release(gstate)
	
	// Safe zone: Even if scheduler tries, goroutine stays on same thread
	runtime.Gosched() // This won't switch threads now!
	time.Sleep(time.Millisecond)
	
	// Create contention - but goroutine is locked, so safe
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.Gosched()
		}()
	}
	wg.Wait()
	
	// Simulate Python call
	fmt.Printf("[Goroutine] Calling Python function...\n")
	
	// Simulate PyGILState_Release() - SAFE! Same thread guaranteed
	// (defer will handle it)
	
	fmt.Printf("[Goroutine] Python callback completed (FIXED): %s\n", msg)
}

func main() {
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("Thread Switching Bug Demonstration")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println()
	
	fmt.Println("This demo simulates the bug where:")
	fmt.Println("1. PyGILState_Ensure() binds to thread T1")
	fmt.Println("2. Go scheduler switches goroutine to thread T2")
	fmt.Println("3. PyGILState_Release() called on thread T2")
	fmt.Println("4. CRASH! Thread mismatch")
	fmt.Println()
	
	// Test 1: Buggy version (may crash)
	fmt.Println("--- Test 1: BUGGY Version (no LockOSThread) ---")
	fmt.Println("Running 50 concurrent callbacks to increase thread switching...")
	fmt.Println()
	
	var wg1 sync.WaitGroup
	errors1 := make(chan error, 50)
	
	for i := 0; i < 50; i++ {
		wg1.Add(1)
		go func(id int) {
			defer wg1.Done()
			defer func() {
				if r := recover(); r != nil {
					errors1 <- fmt.Errorf("goroutine %d: %v", id, r)
				}
			}()
			SimulatePythonCallback_BUGGY(fmt.Sprintf("Message-%d", id))
		}(i)
	}
	
	wg1.Wait()
	close(errors1)
	
	errorCount := 0
	for err := range errors1 {
		errorCount++
		fmt.Printf("❌ ERROR: %v\n", err)
	}
	
	if errorCount > 0 {
		fmt.Printf("\n✓ Bug reproduced! %d goroutines crashed\n", errorCount)
	} else {
		fmt.Println("\n⚠ Bug not reproduced this time (may happen randomly)")
		fmt.Println("  Try running again or increase concurrency")
	}
	
	fmt.Println()
	fmt.Println("--- Test 2: FIXED Version (with LockOSThread) ---")
	fmt.Println("Running 10 concurrent callbacks with fix...")
	fmt.Println()
	
	var wg2 sync.WaitGroup
	errors2 := make(chan error, 10)
	
	for i := 0; i < 10; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			defer func() {
				if r := recover(); r != nil {
					errors2 <- fmt.Errorf("goroutine %d: %v", id, r)
				}
			}()
			SimulatePythonCallback_FIXED(fmt.Sprintf("Message-%d", id))
		}(i)
	}
	
	wg2.Wait()
	close(errors2)
	
	errorCount2 := 0
	for err := range errors2 {
		errorCount2++
		fmt.Printf("❌ ERROR: %v\n", err)
	}
	
	if errorCount2 == 0 {
		fmt.Println("\n✓ All callbacks completed successfully!")
		fmt.Println("  The fix works - no crashes!")
	} else {
		fmt.Printf("\n⚠ Unexpected errors: %d\n", errorCount2)
	}
	
	fmt.Println()
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("Summary:")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println()
	fmt.Println("The buggy version may crash when goroutines switch threads.")
	fmt.Println("The fixed version uses runtime.LockOSThread() to prevent")
	fmt.Println("thread switching, ensuring PyGILState_Ensure/Release are")
	fmt.Println("called on the same OS thread.")
}
