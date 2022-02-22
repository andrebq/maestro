package maestro

type (
	waitChildrenErr struct {
		// TODO: later add the PID of the process
	}
)

func (w waitChildrenErr) Error() string {
	return "maestro: children processes failed to terminate within time"
}
