package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_backup_policy":          ResourceYandexBackupPolicy(),
		"yandex_backup_policy_bindings": ResourceYandexBackupPolicyBindings(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_backup_policy": DataSourceYandexBackupPolicy(),
	}
}
