package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_storage_bucket": ResourceYandexStorageBucket(),
		"yandex_storage_object": ResourceYandexStorageObject(),
	}
}

func DataSources() map[string]*schema.Resource {
	return nil
}
