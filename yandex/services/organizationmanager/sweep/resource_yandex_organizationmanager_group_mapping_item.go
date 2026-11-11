package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func registerResourceYandexOrganizationmanagerGroupMappingItem(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_group_mapping_item", &resource.Sweeper{
		Name: "yandex_organizationmanager_group_mapping_item",
		F:    func(_ string) error { return nil },
		Dependencies: []string{
			"yandex_organizationmanager_saml_federation",
			"yandex_organizationmanager_group",
		},
	})
}
