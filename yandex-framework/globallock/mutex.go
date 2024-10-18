package globallock

import "github.com/yandex-cloud/terraform-provider-yandex/common/mutexkv"

var mutexKV = mutexkv.NewMutexKV()

func GetMutexKV() *mutexkv.MutexKV {
	return mutexKV
}
