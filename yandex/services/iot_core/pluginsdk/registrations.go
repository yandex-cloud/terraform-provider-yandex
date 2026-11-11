package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_iot_core_broker":   ResourceYandexIoTCoreBroker(),
		"yandex_iot_core_device":   ResourceYandexIoTCoreDevice(),
		"yandex_iot_core_registry": ResourceYandexIoTCoreRegistry(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_iot_core_broker":   DataSourceYandexIoTCoreBroker(),
		"yandex_iot_core_device":   DataSourceYandexIoTCoreDevice(),
		"yandex_iot_core_registry": DataSourceYandexIoTCoreRegistry(),
	}
}
