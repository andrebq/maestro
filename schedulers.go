package maestro

import "runtime"

// LoopUntilCancel executes target non-stop until the parent context is
// Done
func LoopUntilCancel(target func(Context) error) func(Context) error {
	return func(c Context) error {
		for {
			select {
			case <-c.Done():
				return c.Err()
			default:
				// let something else run,
				// to prevent an infinite loop from monopolizing
				// the entire processor
				runtime.Gosched()
				target(c)
			}
		}
	}
}
