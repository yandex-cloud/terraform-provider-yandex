package decimal

import "math/big"

type Decimal struct {
	Bytes     [16]byte
	Precision uint32
	Scale     uint32
}

func (d *Decimal) String() string {
	v := FromInt128(d.Bytes, d.Precision, d.Scale)

	return Format(v, d.Precision, d.Scale)
}

func (d *Decimal) BigInt() *big.Int {
	return FromInt128(d.Bytes, d.Precision, d.Scale)
}
