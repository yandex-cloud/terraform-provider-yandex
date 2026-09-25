package meta

import (
	"strconv"

	"google.golang.org/grpc/metadata"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/meta"
)

func ConsumedUnits(md metadata.MD) (consumedUnits uint64) {
	for header, values := range md {
		if header != meta.HeaderConsumedUnits {
			continue
		}
		for _, v := range values {
			v, err := strconv.ParseUint(v, 10, 64)
			if err == nil {
				consumedUnits += v
			}
		}
	}

	return consumedUnits
}
