package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_alb_backend_group": ResourceYandexALBBackendGroup(),
		"yandex_alb_http_router":   ResourceYandexALBHTTPRouter(),
		"yandex_alb_load_balancer": ResourceYandexALBLoadBalancer(),
		"yandex_alb_target_group":  ResourceYandexALBTargetGroup(),
		"yandex_alb_virtual_host":  sdkcommon.AddPassthroughImport(WithALBVirtualHostID(ResourceYandexALBVirtualHost())),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_alb_backend_group": DataSourceYandexALBBackendGroup(),
		"yandex_alb_http_router":   DataSourceYandexALBHTTPRouter(),
		"yandex_alb_load_balancer": DataSourceYandexALBLoadBalancer(),
		"yandex_alb_target_group":  DataSourceYandexALBTargetGroup(),
		"yandex_alb_virtual_host":  DataSourceYandexALBVirtualHost(),
	}
}
