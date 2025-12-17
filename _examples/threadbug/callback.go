// Copyright 2024 The go-python Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package threadbug demonstrates the thread switching bug with Python callbacks
package threadbug

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// CallbackFunc is a function type that can be called from Python
type CallbackFunc func(msg string) string

// CallbackHandler handles Python callbacks
type CallbackHandler struct {
	mu sync.Mutex
}

// CallWithCallback calls the Python callback function
// This function will be called from goroutines that may switch threads
func (h *CallbackHandler) CallWithCallback(cb CallbackFunc, msg string) string {
	if cb == nil {
		return "callback is nil"
	}
	return cb(msg)
}

// TriggerThreadSwitch forces goroutine to switch threads by creating contention
func TriggerThreadSwitch() {
	// Create many goroutines to force thread switching
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Do some work that might cause thread switching
			time.Sleep(time.Millisecond * 10)
			runtime.Gosched() // Explicitly yield to scheduler
		}(i)
	}
	wg.Wait()
}

// GetCurrentThreadID returns a pseudo thread ID for debugging
func GetCurrentThreadID() int64 {
	// Use goroutine ID as a proxy for thread identification
	// Note: This is not a real OS thread ID, but helps demonstrate the issue
	return time.Now().UnixNano() % 1000000
}

// SimulateWebServer simulates a web server scenario with concurrent requests
func SimulateWebServer(cb CallbackFunc, numRequests int) []string {
	results := make([]string, numRequests)
	var wg sync.WaitGroup
	
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Simulate HTTP handler - each runs in its own goroutine
			// This goroutine may run on different OS threads
			
			// Force thread contention to increase chance of thread switching
			TriggerThreadSwitch()
			
			// Now call Python callback - THIS IS WHERE THE BUG OCCURS
			// If goroutine switched threads between PyGILState_Ensure and Release,
			// it will crash!
			handler := &CallbackHandler{}
			result := handler.CallWithCallback(cb, fmt.Sprintf("Request-%d", id))
			results[id] = result
		}(i)
	}
	
	wg.Wait()
	return results
}
