package trace

type sessionInfo interface {
	ID() string
	NodeID() uint32
	Status() string
}
