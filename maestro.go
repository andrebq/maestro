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

		// Triggers a shutdown event but does not block while
		// waiting for the children
		//
		// Useful to shutdown a parent context from a children,
		// provided the children has access to the parent context
		Shutdown()
	}

	base struct {
		ctx context.Context

		children sync.WaitGroup
		cancel   context.CancelFunc
	}

	key byte
)

const (
	maestroCtx = key(1)
)

func New(ctx context.Context) Context {
	ctx, cancel := context.WithCancel(ctx)
	return &base{ctx: ctx, cancel: cancel}
}

func Closest(ctx context.Context) (Context, bool) {
	v := ctx.Value(maestroCtx)
	if v == nil {
		return nil, false
	}
	return v.(Context), true
}

func (b *base) Deadline() (time.Time, bool) { return b.ctx.Deadline() }
func (b *base) Value(key interface{}) interface{} {
	if key == maestroCtx {
		// TODO: should this be a proxy that prevents downstream consumers from messing up with the parent ctx?!
		return b
	}
	return b.ctx.Value(key)
}
func (b *base) Err() error            { return b.ctx.Err() }
func (b *base) Done() <-chan struct{} { return b.ctx.Done() }
func (b *base) Shutdown() {
	b.cancel()
}

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

func SyncShutdown(m Context, timeout <-chan struct{}) error {
	m.Shutdown()
	return m.WaitChildren(timeout)
}
