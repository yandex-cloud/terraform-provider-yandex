package pool

type Stats struct {
	Limit            int
	Index            int
	Idle             int
	Wait             int
	CreateInProgress int
}
