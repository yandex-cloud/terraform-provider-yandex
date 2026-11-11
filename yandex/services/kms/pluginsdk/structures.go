package pluginsdk

import (
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func parseKmsDefaultAlgorithm(algo string) (kms.SymmetricAlgorithm, error) {
	val, ok := kms.SymmetricAlgorithm_value[algo]
	if !ok {
		return kms.SymmetricAlgorithm(0), fmt.Errorf("value for 'default_algorithm' should be one of %s, not `%s`", sdkcommon.GetJoinedKeys(sdkcommon.GetEnumValueMapKeys(kms.SymmetricAlgorithm_value)), algo)
	}
	return kms.SymmetricAlgorithm(val), nil
}
