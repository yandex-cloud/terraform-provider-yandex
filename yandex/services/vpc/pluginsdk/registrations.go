package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_vpc_address":                ResourceYandexVPCAddress(),
		"yandex_vpc_default_security_group": ResourceYandexVPCDefaultSecurityGroup(),
		"yandex_vpc_gateway":                ResourceYandexVPCGateway(),
		"yandex_vpc_network":                ResourceYandexVPCNetwork(),
		"yandex_vpc_route_table":            ResourceYandexVPCRouteTable(),
		"yandex_vpc_security_group":         ResourceYandexVPCSecurityGroup(),
		"yandex_vpc_subnet":                 ResourceYandexVPCSubnet(),
		"yandex_vpc_private_endpoint":       ResourceYandexVPCPrivateEndpoint(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_vpc_address":          DataSourceYandexVPCAddress(),
		"yandex_vpc_gateway":          DataSourceYandexVPCGateway(),
		"yandex_vpc_network":          DataSourceYandexVPCNetwork(),
		"yandex_vpc_route_table":      DataSourceYandexVPCRouteTable(),
		"yandex_vpc_security_group":   DataSourceYandexVPCSecurityGroup(),
		"yandex_vpc_subnet":           DataSourceYandexVPCSubnet(),
		"yandex_vpc_private_endpoint": DataSourceYandexVPCPrivateEndpoint(),
	}
}
