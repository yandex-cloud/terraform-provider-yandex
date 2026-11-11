package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_kms_secret_ciphertext": ResourceYandexKMSSecretCiphertext(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_kms_symmetric_key": DataSourceYandexKMSSymmetricKey(),
	}
}
