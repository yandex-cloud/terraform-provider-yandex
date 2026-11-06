package trino_cluster

import "github.com/yandex-cloud/terraform-provider-yandex/pkg/globallock"

// LockOperations serializes cluster operations, which the API runs one at a time. Take it before starting the timeout
func LockOperations(clusterID string) (unlock func()) {
	key := "yandex_trino_cluster-" + clusterID
	mutexKV := globallock.GetMutexKV()
	mutexKV.Lock(key)
	return func() { mutexKV.Unlock(key) }
}
