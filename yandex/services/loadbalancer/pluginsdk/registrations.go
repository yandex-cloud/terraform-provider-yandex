package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_lb_network_load_balancer": ResourceYandexLBNetworkLoadBalancer(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_lb_network_load_balancer": DataSourceYandexLBNetworkLoadBalancer(),
		"yandex_lb_target_group":          DataSourceYandexLBTargetGroup(),
	}
}
