package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_function":                ResourceYandexFunction(),
		"yandex_function_scaling_policy": ResourceYandexFunctionScalingPolicy(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_function":                DataSourceYandexFunction(),
		"yandex_function_scaling_policy": DataSourceYandexFunctionScalingPolicy(),
	}
}
