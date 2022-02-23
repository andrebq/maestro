# maestro
A simpler way to manage tree of goroutines

## Basic usage

Call `maestro.New(parentCtx)` to acquire a new `maestro.Context` object,
besides the methods available on `context.Context` this object allows
you to spin-up go routines and track their completion.

The advantage of `maestro.Context` is that tracking the lifetime of
child goroutines is easier so you can start a tear-down operation
from the parent context but wait until all children goroutines
have completed their clean-up process.

```go
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	m := New(ctx)
	m.Spawn(func(ctx maestro.Context) error { time.Sleep(time.Second / 2); return nil })
	m.Spawn(func(ctx maestro.Context) error {
		ctx.Spawn(func(ctx maestro.Context) error { time.Sleep(time.Second / 2); return nil })
		// Wait without a timeout
		return ctx.WaitChildren(nil)
	})
	cancel()
	// Process tree have at least 1 second to perform cleanup
	m.WaitChildren(TimeoutAfter(time.Second))
```

The previous code-block shows how easy it is to create a process-tree and wait for all
children processes to finish before returning.
