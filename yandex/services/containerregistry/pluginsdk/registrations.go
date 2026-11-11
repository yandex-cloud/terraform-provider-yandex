package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_container_registry_ip_permission":      ResourceYandexContainerRegistryIPPermission(),
		"yandex_container_repository_lifecycle_policy": ResourceYandexContainerRepositoryLifecyclePolicy(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_container_registry":                    DataSourceYandexContainerRegistry(),
		"yandex_container_registry_ip_permission":      DataSourceYandexContainerRegistryIPPermission(),
		"yandex_container_repository":                  DataSourceYandexContainerRepository(),
		"yandex_container_repository_lifecycle_policy": DataSourceYandexContainerRepositoryLifecyclePolicy(),
	}
}
