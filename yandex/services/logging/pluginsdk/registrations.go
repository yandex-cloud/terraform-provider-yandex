package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_logging_group": DataSourceYandexLoggingGroup(),
	}
}

func Resources() map[string]*schema.Resource {
	return nil
}
