package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_resourcemanager_folder_iam_policy": ResourceYandexResourceManagerFolderIAMPolicy(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_resourcemanager_cloud":  DataSourceYandexResourceManagerCloud(),
		"yandex_resourcemanager_folder": DataSourceYandexResourceManagerFolder(),
	}
}
