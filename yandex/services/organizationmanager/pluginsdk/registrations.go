package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_organizationmanager_saml_federation":              ResourceYandexOrganizationManagerSamlFederation(),
		"yandex_organizationmanager_saml_federation_user_account": ResourceYandexOrganizationManagerSamlFederationUserAccount(),
		"yandex_organizationmanager_group_mapping":                ResourceYandexOrganizationManagerGroupMapping(),
		"yandex_organizationmanager_group_mapping_item":           ResourceYandexOrganizationManagerGroupMappingItem(),
		"yandex_organizationmanager_group_membership":             ResourceYandexOrganizationManagerGroupMembership(),
		"yandex_organizationmanager_os_login_settings":            ResourceYandexOrganizationManagerOsLoginSettings(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_organizationmanager_group":                        DataSourceYandexOrganizationManagerGroup(),
		"yandex_organizationmanager_os_login_settings":            DataSourceYandexOrganizationManagerOsLoginSettings(),
		"yandex_organizationmanager_saml_federation":              DataSourceYandexOrganizationManagerSamlFederation(),
		"yandex_organizationmanager_saml_federation_user_account": DataSourceYandexOrganizationManagerSamlFederationUserAccount(),
	}
}
