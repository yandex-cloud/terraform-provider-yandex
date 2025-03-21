package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
)

func dataSourceYandexOrganizationManagerSamlFederationUserAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a user of Yandex SAML Federation. For more information, see [the official documentation](https://yandex.cloud/docs/organization/operations/federations/integration-common).\n\n~> If terraform user had sufficient access and user specified in data source did not exist, it would be created. This behavior will was **fixed**. Use resource `yandex_organizationmanager_saml_federation_user_account` to manage account lifecycle.\n",

		Read: dataSourceYandexOrganizationManagerSamlFederationUserAccountRead,
		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerSamlFederationUserAccount().Schema["federation_id"].Description,
				Required:    true,
			},
			"name_id": {
				Type:         schema.TypeString,
				Description:  resourceYandexOrganizationManagerSamlFederationUserAccount().Schema["name_id"].Description,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func dataSourceYandexOrganizationManagerSamlFederationUserAccountRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	federationID := d.Get("federation_id").(string)
	nameID := d.Get("name_id").(string)
	req := &saml.ListFederatedUserAccountsRequest{
		FederationId: federationID,
		Filter:       fmt.Sprintf("name_id=%q", nameID),
	}

	listResp, err := config.sdk.OrganizationManagerSAML().Federation().ListUserAccounts(
		config.Context(),
		req,
	)
	if err != nil {
		return err
	}
	if len(listResp.UserAccounts) != 1 {
		return fmt.Errorf("Failed to resolve data source saml user account %s in saml federation %s", nameID, federationID)
	}

	userAccount := listResp.UserAccounts[0]
	d.SetId(userAccount.Id)

	return nil
}
