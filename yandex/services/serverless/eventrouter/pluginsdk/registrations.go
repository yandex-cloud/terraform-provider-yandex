package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_serverless_eventrouter_connector": ResourceYandexServerlessEventrouterConnector(),
		"yandex_serverless_eventrouter_rule":      ResourceYandexServerlessEventrouterRule(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_serverless_eventrouter_bus":       DataSourceYandexServerlessEventrouterBus(),
		"yandex_serverless_eventrouter_connector": DataSourceYandexServerlessEventrouterConnector(),
		"yandex_serverless_eventrouter_rule":      DataSourceYandexServerlessEventrouterRule(),
	}
}
