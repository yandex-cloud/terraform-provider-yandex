package conn

type State int8

const (
	Unknown = State(iota)
	Created
	Online
	Banned
	Offline
	Destroyed
)

func (s State) Code() int {
	return int(s)
}

func (s State) String() string {
	switch s {
	case Created:
		return "created"
	case Online:
		return "online"
	case Banned:
		return "banned"
	case Offline:
		return "offline"
	case Destroyed:
		return "destroyed"
	default:
		return "unknown"
	}
}

func (s State) IsValid() bool {
	switch s {
	case Online, Offline, Banned:
		return true
	default:
		return false
	}
}
