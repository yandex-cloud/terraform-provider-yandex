package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_cdn_origin_group": ResourceYandexCDNOriginGroup(),
		"yandex_cdn_resource":     ResourceYandexCDNResource(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_cdn_origin_group": DataSourceYandexCDNOriginGroup(),
		"yandex_cdn_resource":     DataSourceYandexCDNResource(),
	}
}
