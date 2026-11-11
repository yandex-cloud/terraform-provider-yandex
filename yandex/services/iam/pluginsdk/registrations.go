package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_iam_oauth_client_secret":               ResourceYandexIAMOAuthClientSecret(),
		"yandex_iam_service_account_api_key":           ResourceYandexIAMServiceAccountAPIKey(),
		"yandex_iam_service_account_iam_policy":        ResourceYandexIAMServiceAccountIAMPolicy(),
		"yandex_iam_service_account_key":               ResourceYandexIAMServiceAccountKey(),
		"yandex_iam_service_account_static_access_key": ResourceYandexIAMServiceAccountStaticAccessKey(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_iam_policy":                            DataSourceYandexIAMPolicy(),
		"yandex_iam_role":                              DataSourceYandexIAMRole(),
		"yandex_iam_service_account":                   DataSourceYandexIAMServiceAccount(),
		"yandex_iam_service_agent":                     DataSourceYandexIamServiceAgent(),
		"yandex_iam_user":                              DataSourceYandexIAMUser(),
		"yandex_iam_workload_identity_oidc_federation": DataSourceYandexIAMWorkloadIdentityOidcFederation(),
	}
}
