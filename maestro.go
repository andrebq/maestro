package maestro

import (
	"context"
	"sync"
	"time"
)

type (
	Signal  struct{}
	Done    chan Signal
	Context interface {
		context.Context

		Spawn(func(ctx Context) error)
		WaitChildren(<-chan struct{}) error
	}

	base struct {
		ctx context.Context

		children sync.WaitGroup
	}
)

func New(ctx context.Context) Context {
	return &base{ctx: ctx}
}

func (b *base) Deadline() (time.Time, bool) { return b.ctx.Deadline() }
func (b *base) Value(key interface{}) interface{} {
	return b.ctx.Value(key)
}
func (b *base) Err() error            { return b.ctx.Err() }
func (b *base) Done() <-chan struct{} { return b.ctx.Done() }

func (b *base) Spawn(fn func(ctx Context) error) {
	childCtx := New(b)
	b.children.Add(1)
	go func() {
		defer b.children.Done()
		fn(childCtx)
	}()
}

func (b *base) WaitChildren(timeout <-chan struct{}) error {
	cleanExit := make(chan struct{})
	go func() {
		b.children.Wait()
		close(cleanExit)
	}()
	select {
	case <-cleanExit:
		return nil
	case <-timeout:
		return waitChildrenErr{}
	}
}

func TimeoutAfter(delay time.Duration) <-chan struct{} {
	ch := make(chan struct{})
	time.AfterFunc(delay, func() { close(ch) })
	return ch
}
