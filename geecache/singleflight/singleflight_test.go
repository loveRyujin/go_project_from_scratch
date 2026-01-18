package singleflight

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	var g Group
	v, err := g.Do("key", func() (any, error) {
		return "bar", nil
	})
	if err != nil {
		t.Fatalf("Do error: %v", err)
	}
	if v.(string) != "bar" {
		t.Fatalf("got %v, want bar", v)
	}
}

func TestDoErr(t *testing.T) {
	var g Group
	someErr := errors.New("some error")
	v, err := g.Do("key", func() (any, error) {
		return nil, someErr
	})
	if err != someErr {
		t.Fatalf("Do error: %v, want %v", err, someErr)
	}
	if v != nil {
		t.Fatalf("got %v, want nil", v)
	}
}

func TestDoDupSuppress(t *testing.T) {
	var g Group
	c := make(chan string)
	var calls int32
	fn := func() (any, error) {
		atomic.AddInt32(&calls, 1)
		return <-c, nil
	}

	const n = 10
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			v, err := g.Do("key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
			}
			if v.(string) != "bar" {
				t.Errorf("got %v, want bar", v)
			}
		})
	}
	time.Sleep(100 * time.Millisecond) // 让所有 goroutine 都进入 Do
	c <- "bar"
	wg.Wait()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("number of calls = %d, want 1", got)
	}
}

func TestDoDifferentKeys(t *testing.T) {
	var g Group
	var calls int32
	fn := func() (any, error) {
		atomic.AddInt32(&calls, 1)
		return "result", nil
	}

	const n = 10
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			key := "key" + string(rune('0'+i))
			v, err := g.Do(key, fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
			}
			if v.(string) != "result" {
				t.Errorf("got %v, want result", v)
			}
		})
	}
	wg.Wait()
	if got := atomic.LoadInt32(&calls); got != n {
		t.Errorf("number of calls = %d, want %d", got, n)
	}
}

func TestDoConcurrent(t *testing.T) {
	var g Group
	var calls int32
	fn := func() (any, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(10 * time.Millisecond)
		return "result", nil
	}

	const n = 100
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			v, err := g.Do("key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
			}
			if v.(string) != "result" {
				t.Errorf("got %v, want result", v)
			}
		})
	}
	wg.Wait()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("number of calls = %d, want 1", got)
	}
}

func TestDoErrorPropagation(t *testing.T) {
	var g Group
	someErr := errors.New("some error")
	c := make(chan struct{})
	var calls int32
	fn := func() (any, error) {
		atomic.AddInt32(&calls, 1)
		<-c // 等待所有 goroutine 都进入 Do
		return nil, someErr
	}

	const n = 10
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			v, err := g.Do("key", fn)
			if err != someErr {
				t.Errorf("Do error: %v, want %v", err, someErr)
			}
			if v != nil {
				t.Errorf("got %v, want nil", v)
			}
		})
	}
	time.Sleep(100 * time.Millisecond) // 让所有 goroutine 都进入 Do
	close(c)
	wg.Wait()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("number of calls = %d, want 1", got)
	}
}

func TestDoCleanup(t *testing.T) {
	var g Group
	fn := func() (any, error) {
		return "result", nil
	}

	// 第一次调用
	v1, err1 := g.Do("key", fn)
	if err1 != nil {
		t.Fatalf("Do error: %v", err1)
	}
	if v1.(string) != "result" {
		t.Fatalf("got %v, want result", v1)
	}

	// 检查 map 是否已清理
	g.mu.Lock()
	if _, ok := g.m["key"]; ok {
		t.Fatal("key should be removed from map after Do completes")
	}
	g.mu.Unlock()

	// 第二次调用应该重新执行函数
	var calls int32
	fn2 := func() (any, error) {
		atomic.AddInt32(&calls, 1)
		return "result2", nil
	}
	v2, err2 := g.Do("key", fn2)
	if err2 != nil {
		t.Fatalf("Do error: %v", err2)
	}
	if v2.(string) != "result2" {
		t.Fatalf("got %v, want result2", v2)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("number of calls = %d, want 1", got)
	}
}

func TestDoNilMap(t *testing.T) {
	var g Group
	// g.m 初始为 nil，应该自动初始化
	v, err := g.Do("key", func() (any, error) {
		return "result", nil
	})
	if err != nil {
		t.Fatalf("Do error: %v", err)
	}
	if v.(string) != "result" {
		t.Fatalf("got %v, want result", v)
	}
}
