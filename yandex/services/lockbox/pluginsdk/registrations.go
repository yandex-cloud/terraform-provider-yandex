package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_lockbox_secret":                ResourceYandexLockboxSecret(),
		"yandex_lockbox_secret_version":        ResourceYandexLockboxSecretVersion(),
		"yandex_lockbox_secret_version_hashed": ResourceYandexLockboxSecretVersionHashed(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_lockbox_secret":               DataSourceYandexLockboxSecret(),
		"yandex_lockbox_secret_version":       DataSourceYandexLockboxSecretVersion(),
		"yandex_lockbox_secret_version_entry": DataSourceYandexLockboxSecretVersionEntry(),
	}
}
