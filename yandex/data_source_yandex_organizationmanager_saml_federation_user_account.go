package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
)

func dataSourceYandexOrganizationManagerSamlFederationUserAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexOrganizationManagerSamlFederationUserAccountRead,
		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name_id": {
				Type:         schema.TypeString,
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
	var nextPageToken string
	for {
		req := &saml.ListFederatedUserAccountsRequest{
			FederationId: federationID,
		}

		if nextPageToken != "" {
			req.PageToken = nextPageToken
		}

		listResp, err := config.sdk.OrganizationManagerSAML().Federation().ListUserAccounts(
			config.Context(),
			req,
		)
		if err != nil {
			return err
		}

		for _, account := range listResp.UserAccounts {
			if account.GetSamlUserAccount().GetNameId() == nameID {
				d.SetId(account.Id)
				return nil
			}
		}

		if listResp.NextPageToken == "" {
			break
		}

		nextPageToken = listResp.NextPageToken
	}

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManagerSAML().Federation().AddUserAccounts(
		config.Context(),
		&saml.AddFederatedUserAccountsRequest{
			FederationId: federationID,
			NameIds:      []string{nameID},
		},
	))

	if err != nil {
		return err
	}

	err = op.Wait(config.Context())
	if err != nil {
		return err
	}

	rawResp, err := op.Response()
	if err != nil {
		return err
	}
	addResp := rawResp.(*saml.AddFederatedUserAccountsResponse)
	d.SetId(addResp.UserAccounts[0].Id)

	return nil
}
