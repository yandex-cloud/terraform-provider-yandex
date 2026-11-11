package sdkcommon

import (
	"fmt"
)

func InvalidKeyError(name string, valid map[string]int32, got string) error {
	keys := make([]string, 0, len(valid))
	for k := range valid {
		keys = append(keys, k)
	}
	return fmt.Errorf("expected %q to be one of %v, got %q", name, keys, got)
}
