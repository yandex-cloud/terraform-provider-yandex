package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func registerResourceYandexOrganizationmanagerSamlFederationUserAccount(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_saml_federation_user_account", &resource.Sweeper{
		Name:         "yandex_organizationmanager_saml_federation_user_account",
		F:            func(_ string) error { return nil },
		Dependencies: []string{"yandex_organizationmanager_saml_federation"},
	})
}
