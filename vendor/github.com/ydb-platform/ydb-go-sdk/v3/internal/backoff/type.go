package backoff

import "fmt"

// Type reports how to Backoff operation
type Type uint8

// Binary flags that used as Type
const (
	TypeNoBackoff Type = 1 << iota >> 1

	TypeFast
	TypeSlow

	TypeAny = TypeFast | TypeSlow
)

func (b Type) String() string {
	switch b {
	case TypeNoBackoff:
		return "immediately"
	case TypeFast:
		return "fast backoff"
	case TypeSlow:
		return "slow backoff"
	case TypeAny:
		return "any backoff"
	default:
		return fmt.Sprintf("unknown backoff type %d", b)
	}
}
