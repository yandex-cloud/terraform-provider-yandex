package trino_access_control

import (
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/trino_cluster"
)

const (
	// All modifying actions on access control leads to ClusterService.Update call.
	YandexTrinoAccessControlCreateTimeout = trino_cluster.YandexTrinoClusterUpdateTimeout
	YandexTrinoAccessControlUpdateTimeout = YandexTrinoAccessControlCreateTimeout
	YandexTrinoAccessControlDeleteTimeout = YandexTrinoAccessControlCreateTimeout
)
