#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
真实测试脚本 - 复现线程切换 bug

这个脚本使用真实的 gopy 绑定来测试 bug。
需要先构建绑定：gopy build -output=./out -vm=python3 ...
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
    print("  cd out")
    print("  python3 ../test_bug.py")
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
        return True
    except Exception as e:
        print(f"\n✗ CRASHED with error: {e}")
        print(f"Error type: {type(e).__name__}")
        import traceback
        traceback.print_exc()
        return False

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
        return True
    except Exception as e:
        print(f"\n✗ CRASHED with error: {e}")
        print(f"Error type: {type(e).__name__}")
        import traceback
        traceback.print_exc()
        return False

def main():
    print("=" * 70)
    print("Thread Switching Bug - 真实测试")
    print("=" * 70)
    print("\n这个测试使用真实的 gopy 绑定来复现 bug：")
    print("1. Go goroutine 调用 PyGILState_Ensure() 在线程 T1")
    print("2. Go 调度器将 goroutine 切换到线程 T2")
    print("3. PyGILState_Release() 在线程 T2 上调用")
    print("4. CRASH! 线程不匹配导致 segfault")
    print("\n" + "=" * 70)
    
    # Test 1: Basic callback (should work)
    test_basic_callback()
    
    # Test 2: Concurrent callbacks (may crash)
    success = test_concurrent_callbacks()
    
    if success:
        # Test 3: High contention (most likely to crash)
        test_high_contention()
    
    print("\n" + "=" * 70)
    print("测试完成！")
    print("=" * 70)
    print("\n如果测试崩溃了，说明成功复现了 bug！")
    print("修复方法：在 gopy 生成的代码中添加 runtime.LockOSThread()")

if __name__ == "__main__":
    main()
