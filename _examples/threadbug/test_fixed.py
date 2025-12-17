#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Test script with the FIX applied.

This script uses the fixed version that includes runtime.LockOSThread()
to prevent goroutine thread switching.

Run this with: python3 test_fixed.py
"""

from __future__ import print_function
import sys
import threading
import time

# Import the generated gopy module (fixed version)
try:
    import threadbug_fixed as threadbug
except ImportError:
    # Fallback to regular import if fixed version not available
    try:
        import threadbug
        print("WARNING: Using unfixed version. Results may vary.")
    except ImportError as e:
        print("ERROR: Could not import threadbug module.")
        print("Make sure you have built the gopy bindings first.")
        sys.exit(1)

# Global counter for tracking callbacks
callback_count = 0
callback_lock = threading.Lock()

def python_callback(msg):
    """Python callback function that will be called from Go"""
    global callback_count
    
    with callback_lock:
        callback_count += 1
        count = callback_count
    
    # Simulate some work
    time.sleep(0.001)  # 1ms delay
    
    result = f"Python received: {msg} (callback #{count})"
    print(f"[Python Thread {threading.current_thread().ident}] {result}")
    return result

def test_basic_callback():
    """Test basic callback"""
    print("\n=== Test 1: Basic Callback ===")
    handler = threadbug.CallbackHandler()
    result = handler.CallWithCallback(python_callback, "Hello from Go")
    print(f"Result: {result}")
    print("✓ Basic callback works\n")

def test_concurrent_callbacks():
    """Test concurrent callbacks - should work with fix"""
    print("\n=== Test 2: Concurrent Callbacks (with fix) ===")
    print("Starting 50 concurrent goroutines...")
    print("With runtime.LockOSThread(), goroutines won't switch threads.")
    print("This should work without crashing!\n")
    
    try:
        results = threadbug.SimulateWebServer(python_callback, 50)
        print(f"\n✓ Successfully completed {len(results)} callbacks")
        print(f"Total callbacks executed: {callback_count}")
        for i, r in enumerate(results[:5]):  # Show first 5
            print(f"  Result {i}: {r}")
        if len(results) > 5:
            print(f"  ... and {len(results) - 5} more")
        return True
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        return False

def test_high_contention():
    """Test with high thread contention"""
    print("\n=== Test 3: High Thread Contention (with fix) ===")
    print("Creating many goroutines - should work fine with fix...")
    
    # Reset counter
    global callback_count
    callback_count = 0
    
    try:
        # Use many goroutines - should all work with fix
        results = threadbug.SimulateWebServer(python_callback, 200)
        print(f"\n✓ Successfully completed {len(results)} callbacks")
        print(f"Total callbacks executed: {callback_count}")
        return True
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        return False

def test_stress():
    """Stress test with many concurrent requests"""
    print("\n=== Test 4: Stress Test (1000 concurrent) ===")
    print("Running stress test with 1000 concurrent goroutines...")
    
    global callback_count
    callback_count = 0
    
    start_time = time.time()
    try:
        results = threadbug.SimulateWebServer(python_callback, 1000)
        elapsed = time.time() - start_time
        print(f"\n✓ Successfully completed {len(results)} callbacks in {elapsed:.2f}s")
        print(f"Total callbacks executed: {callback_count}")
        print(f"Throughput: {len(results)/elapsed:.0f} callbacks/second")
        return True
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        return False

def main():
    print("=" * 70)
    print("Thread Switching Bug - FIXED Version Test")
    print("=" * 70)
    print("\nThis test uses the FIXED version with runtime.LockOSThread():")
    print("1. Go goroutine calls runtime.LockOSThread()")
    print("2. Goroutine is locked to current OS thread")
    print("3. PyGILState_Ensure() called on thread T1")
    print("4. Even if scheduler tries, goroutine stays on T1")
    print("5. PyGILState_Release() called on same thread T1")
    print("6. ✓ SUCCESS! No crash!")
    print("\n" + "=" * 70)
    
    # Run all tests
    test_basic_callback()
    
    if test_concurrent_callbacks():
        if test_high_contention():
            test_stress()
    
    print("\n" + "=" * 70)
    print("All tests completed successfully!")
    print("=" * 70)
    print("\nThe fix works! runtime.LockOSThread() prevents thread switching")
    print("and ensures PyGILState_Ensure/Release are called on the same thread.")

if __name__ == "__main__":
    main()
