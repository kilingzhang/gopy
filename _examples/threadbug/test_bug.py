#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Test script to reproduce the thread switching bug.

This script demonstrates the crash that occurs when:
1. Go goroutine calls Python callback
2. Goroutine switches OS threads between PyGILState_Ensure and Release
3. PyGILState_Release crashes because it's called from wrong thread

Run this with: python3 test_bug.py
"""

from __future__ import print_function
import sys
import threading
import time

# Import the generated gopy module
try:
    import threadbug
except ImportError as e:
    print("ERROR: Could not import threadbug module.")
    print("Make sure you have built the gopy bindings first:")
    print("  cd _examples/threadbug")
    print("  gopy build -output=./out -vm=python3 github.com/go-python/gopy/_examples/threadbug")
    print("  python3 -c 'import sys; sys.path.insert(0, \"out\"); import threadbug'")
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
    """Test basic callback - should work fine"""
    print("\n=== Test 1: Basic Callback (should work) ===")
    handler = threadbug.CallbackHandler()
    result = handler.CallWithCallback(python_callback, "Hello from Go")
    print(f"Result: {result}")
    print("✓ Basic callback works\n")

def test_concurrent_callbacks():
    """Test concurrent callbacks - may trigger thread switching bug"""
    print("\n=== Test 2: Concurrent Callbacks (may crash) ===")
    print("Starting 50 concurrent goroutines...")
    print("Each goroutine will call Python callback.")
    print("If goroutines switch threads, this may CRASH!\n")
    
    try:
        results = threadbug.SimulateWebServer(python_callback, 50)
        print(f"\n✓ Successfully completed {len(results)} callbacks")
        print(f"Total callbacks executed: {callback_count}")
        for i, r in enumerate(results[:5]):  # Show first 5
            print(f"  Result {i}: {r}")
        if len(results) > 5:
            print(f"  ... and {len(results) - 5} more")
    except Exception as e:
        print(f"\n✗ CRASHED with error: {e}")
        print(f"Error type: {type(e).__name__}")
        import traceback
        traceback.print_exc()
        return False
    
    return True

def test_high_contention():
    """Test with high thread contention to force thread switching"""
    print("\n=== Test 3: High Thread Contention (most likely to crash) ===")
    print("Creating many goroutines with explicit thread switching...")
    
    # Reset counter
    global callback_count
    callback_count = 0
    
    try:
        # Use more goroutines to increase chance of thread switching
        results = threadbug.SimulateWebServer(python_callback, 200)
        print(f"\n✓ Successfully completed {len(results)} callbacks")
        print(f"Total callbacks executed: {callback_count}")
    except Exception as e:
        print(f"\n✗ CRASHED with error: {e}")
        print(f"Error type: {type(e).__name__}")
        import traceback
        traceback.print_exc()
        return False
    
    return True

def main():
    print("=" * 70)
    print("Thread Switching Bug Reproduction Test")
    print("=" * 70)
    print("\nThis test demonstrates the bug where:")
    print("1. Go goroutine calls PyGILState_Ensure() on thread T1")
    print("2. Go scheduler switches goroutine to thread T2")
    print("3. PyGILState_Release() called on thread T2")
    print("4. CRASH! Thread mismatch causes segfault")
    print("\n" + "=" * 70)
    
    # Test 1: Basic callback (should work)
    test_basic_callback()
    
    # Test 2: Concurrent callbacks (may crash)
    success = test_concurrent_callbacks()
    
    if success:
        # Test 3: High contention (most likely to crash)
        test_high_contention()
    
    print("\n" + "=" * 70)
    print("Test completed!")
    print("=" * 70)
    print("\nNote: If the test crashed, you've successfully reproduced the bug!")
    print("The fix is to use runtime.LockOSThread() before PyGILState_Ensure()")

if __name__ == "__main__":
    main()
