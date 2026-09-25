package secret

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

func Password(password string) string {
	var (
		bytes         = []byte(password)
		startPosition = 3
		endPosition   = len(bytes) - 2
	)
	if startPosition > endPosition {
		for i := range bytes {
			bytes[i] = '*'
		}
	} else {
		for i := startPosition; i < endPosition; i++ {
			bytes[i] = '*'
		}
	}

	return xstring.FromBytes(bytes)
}
