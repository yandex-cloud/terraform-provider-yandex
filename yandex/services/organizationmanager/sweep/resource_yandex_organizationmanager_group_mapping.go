package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func registerResourceYandexOrganizationmanagerGroupMapping(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_group_mapping", &resource.Sweeper{
		Name:         "yandex_organizationmanager_group_mapping",
		F:            func(_ string) error { return nil },
		Dependencies: []string{"yandex_organizationmanager_saml_federation"},
	})
}
