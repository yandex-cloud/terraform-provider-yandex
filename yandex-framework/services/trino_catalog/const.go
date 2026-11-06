package trino_catalog

import "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/trino_cluster"

const (
	// A catalog change restarts the whole cluster.
	YandexTrinoCatalogCreateTimeout = trino_cluster.YandexTrinoClusterUpdateTimeout
	YandexTrinoCatalogDeleteTimeout = trino_cluster.YandexTrinoClusterUpdateTimeout
	YandexTrinoCatalogUpdateTimeout = trino_cluster.YandexTrinoClusterUpdateTimeout

	PasswordStubOnImport = "<real value unknown because resource was imported>"
)
