# Decimal package

Package decimal provides tools for working with YDB's decimal types.

Decimal types are stored as int128 values inside YDB and represented as 16-byte
arrays in ydb package and as *math/big.Int in ydb/decimal package.

Note that returned big.Int values are scaled. That is, math operations must be
prepared keeping in mind scaling factor:

	import (
		"ydb/decimal"
		"math/big"
	)

	var scaleFactor = big.NewInt(10000000000) // Default scale is 9.

	func main() {
		x := decimal.FromInt128([16]byte{...})
		x.Add(x, big.NewInt(42)) // Incorrect.
		x.Add(x, scale(42))      // Correct.
	}

	func scale(n int64) *big.Int {
		x := big.NewInt(n)
		return x.Mul(x, scaleFactor)
	}
