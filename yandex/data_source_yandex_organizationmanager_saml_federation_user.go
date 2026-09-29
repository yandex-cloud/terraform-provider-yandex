package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func dataSourceYandexOrganizationManagerSamlFederationUser() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a user of Yandex SAML Federation by their IAM user account ID. For more information, see [the official documentation](https://yandex.cloud/docs/organization/operations/federations/integration-common).\n",

		Read: dataSourceYandexOrganizationManagerSamlFederationUserRead,
		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:         schema.TypeString,
				Description:  "ID of a SAML Federation.",
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"user_account_id": {
				Type:         schema.TypeString,
				Description:  "ID of the IAM user account.",
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name ID of the SAML federated user.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Full name of the user, as provided by the identity provider.",
			},
			"given_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Given (first) name of the user, as provided by the identity provider.",
			},
			"family_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Family (last) name of the user, as provided by the identity provider.",
			},
			"preferred_username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username of the user, as provided by the identity provider.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email of the user, as provided by the identity provider.",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Additional SAML attributes of the user, as sent by the identity provider in the SAML assertion. If an attribute has multiple values, they are joined with a comma.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceYandexOrganizationManagerSamlFederationUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	federationID := d.Get("federation_id").(string)
	userAccountID := d.Get("user_account_id").(string)

	userAccount, err := config.sdk.IAM().UserAccount().Get(ctx, &iam.GetUserAccountRequest{
		UserAccountId: userAccountID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("user account with ID %q", userAccountID))
	}

	samlUserAccount := userAccount.GetSamlUserAccount()
	if samlUserAccount == nil {
		return fmt.Errorf("user account %q is not a SAML federation user account", userAccountID)
	}
	if samlUserAccount.FederationId != federationID {
		return fmt.Errorf("user account %q belongs to saml federation %q, not %q", userAccountID, samlUserAccount.FederationId, federationID)
	}

	nameID := samlUserAccount.NameId
	if err := d.Set("name_id", nameID); err != nil {
		return err
	}

	attributes := make(map[string]string, len(samlUserAccount.GetAttributes()))
	for key, attr := range samlUserAccount.GetAttributes() {
		attributes[key] = strings.Join(attr.GetValue(), ",")
	}
	if err := d.Set("attributes", attributes); err != nil {
		return err
	}

	user, err := getSamlUserAccount(ctx, config, federationID, nameID)
	if err != nil {
		return fmt.Errorf("error reading saml user '%s': %s", nameID, err)
	}

	claims := user.GetSubjectClaims()
	if err := d.Set("name", claims.GetName()); err != nil {
		return err
	}
	if err := d.Set("given_name", claims.GetGivenName()); err != nil {
		return err
	}
	if err := d.Set("family_name", claims.GetFamilyName()); err != nil {
		return err
	}
	if err := d.Set("preferred_username", claims.GetPreferredUsername()); err != nil {
		return err
	}
	if err := d.Set("email", claims.GetEmail()); err != nil {
		return err
	}

	d.SetId(userAccount.Id)

	return nil
}
