package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_dns_recordset": ResourceYandexDnsRecordSet(),
		"yandex_dns_zone":      ResourceYandexDnsZone(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_dns_zone": DataSourceYandexDnsZone(),
	}
}
