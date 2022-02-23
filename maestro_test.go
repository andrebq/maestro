package maestro

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestSpawner(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var count int32

	m := New(ctx)
	m.Spawn(func(ctx Context) error {
		for i := 0; i < 10; i++ {
			ctx.Spawn(func(ctx Context) error {
				atomic.AddInt32(&count, 1)
				return nil
			})
		}
		// nil blocks forever
		return ctx.WaitChildren(nil)
	})
	err := m.WaitChildren(nil)
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&count) != 10 {
		t.Fatal("Exit before the expected number of processes had finished")
	}
}

func TestTerminationTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	m := New(ctx)
	m.Spawn(func(ctx Context) error {
		time.Sleep(time.Second * 10)
		return nil
	})
	err := m.WaitChildren(TimeoutAfter(time.Millisecond * 100))
	if !errors.Is(err, waitChildrenErr{}) {
		t.Fatalf("Not the expected error: %v", err)
	}
}

func TestLoopUntilCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	m := New(ctx)
	var count int
	var ctxCancel bool
	stop := make(chan struct{})
	m.Spawn(LoopUntilCancel(func(ctx Context) error {
		count++
		if count <= 2 {
			return errors.New("forcing an error")
		}
		select {
		case <-ctx.Done():
			ctxCancel = true
		case <-stop:
		}
		return nil
	}))
	for count < 2 {
		runtime.Gosched()
		select {
		case <-ctx.Done():
			t.Fatal("failed without reaching the expected count")
		default:
		}
	}
	cancel()
	m.WaitChildren(TimeoutAfter(time.Second))
	close(stop)
	if !ctxCancel {
		t.Fatal("The children should have received the cancel call")
	}
}

func TestWaitChildreIsNotImmediate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	childrenDone := make(Done)
	m := New(ctx)
	m.Spawn(func(ctx Context) error {
		<-ctx.Done()
		close(childrenDone)
		return nil
	})
	go func() {
		// parent context got cancel by
		// an external entity
		cancel()
	}()
	m.WaitChildren(nil)
	select {
	case <-childrenDone:
	default:
		t.Fatal("WaitChildren should wait until all childrens are done")
	}
}

func BenchmarkSpawnChildren(b *testing.B) {
	b.StopTimer()
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	newChildren := func(ctx Context) func(ctx Context) error {
		return func(ctx Context) error {
			select {
			case <-ctx.Done():
			case <-rootCtx.Done():
			}
			return nil
		}
	}
	m := New(rootCtx)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Spawn(newChildren(m))
	}
	cancel()
	m.WaitChildren(nil)
}

func BenchmarkNestedSpawn(b *testing.B) {
	b.StopTimer()
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var newChildren func(int) func(Context) error
	newChildren = func(level int) func(ctx Context) error {
		return func(ctx Context) error {
			if level > 0 {
				// spawn new children until level is 0
				ctx.Spawn(newChildren(level - 1))
			}
			select {
			case <-ctx.Done():
			case <-rootCtx.Done():
			}
			ctx.WaitChildren(nil)
			return nil
		}
	}
	m := New(rootCtx)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// each new root children will spawn at least 10 other nested children
		// and wait for all of them to finish
		m.Spawn(newChildren(10))
	}
	cancel()
	m.WaitChildren(nil)
}
