package sdkcommon

import "github.com/yandex-cloud/terraform-provider-yandex/common/mutexkv"

// MutexKV preserves the original provider-wide lock scope across services.
var MutexKV = mutexkv.NewMutexKV()
