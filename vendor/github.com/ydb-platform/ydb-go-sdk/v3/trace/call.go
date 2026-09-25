package trace

import "fmt"

type call interface {
	fmt.Stringer
}
