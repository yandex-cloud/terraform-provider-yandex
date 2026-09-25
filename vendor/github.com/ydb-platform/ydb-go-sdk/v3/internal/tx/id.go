package tx

var _ Identifier = LazyID{}

const (
	LazyTxID = "LAZY_TX"
	FakeTxID = "FAKE_TX"
)

type (
	Identifier interface {
		ID() string
		isYdbTx()
	}
	LazyID struct {
		v *string
	}
)

func (id LazyID) ID() string {
	if id.v == nil {
		return LazyTxID
	}

	return *id.v
}

func (id *LazyID) SetTxID(txID string) {
	id.v = &txID
}

func (id LazyID) isYdbTx() {}

func ID(id string) LazyID {
	return LazyID{v: &id}
}
