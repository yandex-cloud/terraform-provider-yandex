package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_kubernetes_cluster":    ResourceYandexKubernetesCluster(),
		"yandex_kubernetes_node_group": ResourceYandexKubernetesNodeGroup(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_kubernetes_cluster":    DataSourceYandexKubernetesCluster(),
		"yandex_kubernetes_node_group": DataSourceYandexKubernetesNodeGroup(),
	}
}
