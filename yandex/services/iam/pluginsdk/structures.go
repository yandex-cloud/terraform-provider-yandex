package pluginsdk

import (
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func parseIamKeyAlgorithm(algorithm string) (iam.Key_Algorithm, error) {
	val, ok := iam.Key_Algorithm_value[algorithm]
	if !ok {
		return iam.Key_ALGORITHM_UNSPECIFIED, fmt.Errorf("value for 'key_algorithm' should be one of %s, not `%s`", sdkcommon.GetJoinedKeys(sdkcommon.GetEnumValueMapKeys(iam.KeyFormat_value)), algorithm)
	}
	return iam.Key_Algorithm(val), nil
}

func parseIamKeyFormat(format string) (iam.KeyFormat, error) {
	val, ok := iam.KeyFormat_value[format]
	if !ok {
		return iam.KeyFormat(0), fmt.Errorf("value for 'format' should be one of %s, not `%s`", sdkcommon.GetJoinedKeys(sdkcommon.GetEnumValueMapKeys(iam.KeyFormat_value)), format)
	}
	return iam.KeyFormat(val), nil
}
