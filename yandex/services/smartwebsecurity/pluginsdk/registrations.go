package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_sws_security_profile_advanced_rate_limiter_profile_attachment": ResourceYandexSWSSecurityProfileAdvancedRateLimiterProfileAttachment(),
		"yandex_sws_security_profile_waf_profile_attachment":                   ResourceYandexSWSSecurityProfileWAFProfileAttachment(),
	}
}

func DataSources() map[string]*schema.Resource {
	return nil
}
