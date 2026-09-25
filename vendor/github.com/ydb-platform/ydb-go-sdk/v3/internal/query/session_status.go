package query

import "fmt"

type Status uint32

const (
	StatusUnknown = Status(iota)
	StatusIdle
	StatusInUse
	StatusClosing
	StatusClosed
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusUnknown:
		return "Unknown"
	case StatusIdle:
		return "Idle"
	case StatusInUse:
		return "InUse"
	case StatusClosing:
		return "Closing"
	case StatusClosed:
		return "Closed"
	case StatusError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown%d", s)
	}
}

func IsAlive(status Status) bool {
	switch status {
	case StatusClosed, StatusClosing, StatusError:
		return false
	default:
		return true
	}
}
