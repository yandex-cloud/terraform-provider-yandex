# MultiChoose

Store multi-choose list state for Go.

## Usage

```go
package main

import (
	"fmt"

	"github.com/cqroot/multichoose"
)

func main() {
	mc := multichoose.New(10)

	// Set the maximum number of items that can be selected
	mc.SetLimit(3)

	mc.Select(1)
	// true
	fmt.Println(mc.IsSelected(1))

	mc.Deselect(1)
	// false
	fmt.Println(mc.IsSelected(1))
}
```
