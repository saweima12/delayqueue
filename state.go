package delaywheel

import "fmt"

type wheelState uint32

const (
	STATE_TERMINATED wheelState = iota
	STATE_READY
	STATE_SHUTTING_DOWN
)

func (ls wheelState) String() string {
	switch ls {
	case STATE_TERMINATED:
		return "TERMINATED"
	case STATE_READY:
		return "READY"
	case STATE_SHUTTING_DOWN:
		return "SHUTTING_DOWN"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(ls))
	}
}
