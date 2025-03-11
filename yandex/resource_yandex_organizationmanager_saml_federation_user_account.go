package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
)

const yandexOrganizationManagerSamlFederationUserDefaultTimeout = 1 * time.Minute

func resourceYandexOrganizationManagerSamlFederationUserAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexOrganizationManagerSamlFederationUserAccountCreate,
		ReadContext:   resourceYandexOrganizationManagerSamlFederationUserAccountRead,
		DeleteContext: resourceYandexOrganizationManagerSamlFederationUserAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceYandexOrganizationManagerSamlFederationUserAccountImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexOrganizationManagerSamlFederationUserDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexOrganizationManagerSamlFederationUserDefaultTimeout),
			Update: schema.DefaultTimeout(yandexOrganizationManagerSamlFederationUserDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexOrganizationManagerSamlFederationUserDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func resourceYandexOrganizationManagerSamlFederationUserAccountImport(context context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)

	req := &iam.GetUserAccountRequest{
		UserAccountId: d.Id(),
	}
	userAccount, err := config.sdk.IAM().UserAccount().Get(context, req)
	if err != nil {
		return nil, handleNotFoundError(err, d, fmt.Sprintf("Saml user account with ID %q", d.Id()))
	}

	samlUserAccount := userAccount.GetSamlUserAccount()
	federationID := samlUserAccount.FederationId
	nameID := samlUserAccount.NameId

	_, err = getSamlUserAccount(context, config, federationID, nameID)
	if err != nil {
		log.Printf("[WARN] Removing %s because resource doesn't exist anymore", nameID)
		d.SetId("")
		return nil, fmt.Errorf("error reading saml user '%s': %s", nameID, err)
	}

	d.Set("name_id", nameID)
	d.Set("federation_id", federationID)

	return []*schema.ResourceData{d}, nil
}

func resourceYandexOrganizationManagerSamlFederationUserAccountCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	federationID, nameID := d.Get("federation_id").(string), d.Get("name_id").(string)
	req := &saml.AddFederatedUserAccountsRequest{
		FederationId: federationID,
		NameIds:      []string{nameID},
	}

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManagerSAML().Federation().AddUserAccounts(config.Context(), req))
	if err != nil {
		return diag.Errorf("error on add user '%s' operation creation  into federation '%s': %s", nameID, federationID, err)
	}

	err = op.Wait(context)
	if err != nil {
		return diag.Errorf("error on add user '%s' operation wait into federation '%s': %s", nameID, federationID, err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.Errorf("error on adding user '%s' into federation '%s': %s", nameID, federationID, err)
	}

	return resourceYandexOrganizationManagerSamlFederationUserAccountRead(context, d, meta)
}

func resourceYandexOrganizationManagerSamlFederationUserAccountRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	federationID, nameID := d.Get("federation_id").(string), d.Get("name_id").(string)
	user, err := getSamlUserAccount(context, config, federationID, nameID)
	if err != nil {
		log.Printf("[WARN] Removing %s because resource doesn't exist anymore", nameID)
		d.SetId("")
		return diag.Errorf("error reading saml user '%s': %s", nameID, err)
	}
	d.SetId(user.GetSubjectClaims().Sub)

	return nil
}

func resourceYandexOrganizationManagerSamlFederationUserAccountDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	federationID, nameID := d.Get("federation_id").(string), d.Get("name_id").(string)

	req := &saml.DeleteFederatedUserAccountsRequest{
		FederationId: federationID,
		SubjectIds:   []string{d.Id()},
	}

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManagerSAML().Federation().DeleteUserAccounts(context, req))
	if err != nil {
		return diag.Errorf("error on delete saml user '%s' operation creation: %s", nameID, err)
	}

	err = op.Wait(context)
	if err != nil {
		return diag.Errorf("error on delete saml user '%s' operation wait: %s", nameID, err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.Errorf("error deleting user '%s': %s", nameID, err)
	}

	d.SetId("")

	return nil
}

func getSamlFederation(context context.Context, config *Config, federationID string) (*saml.Federation, error) {
	getFederationReq := &saml.GetFederationRequest{
		FederationId: federationID,
	}

	federation, err := config.sdk.OrganizationManagerSAML().Federation().Get(context, getFederationReq)
	if err != nil {
		return nil, fmt.Errorf("error on reading federation '%s': %s", federationID, err)
	}

	return federation, nil
}

func getSamlUserAccount(context context.Context, config *Config, federationID, nameID string) (*organizationmanager.ListMembersResponse_OrganizationUser, error) {
	federation, err := getSamlFederation(context, config, federationID)
	if err != nil {
		return nil, fmt.Errorf("error reading saml user '%s': %s", nameID, federationID)
	}

	organizationID := federation.OrganizationId

	var nextPageToken string
	for {
		req := &organizationmanager.ListMembersRequest{
			OrganizationId: organizationID,
			PageToken:      nextPageToken,
		}

		listResp, err := config.sdk.OrganizationManager().User().ListMembers(context, req)
		if err != nil {
			return nil, fmt.Errorf("error on listing members in organization '%s': %s", organizationID, err)
		}
		for _, account := range listResp.Users {
			if account.SubjectClaims.PreferredUsername == nameID &&
				account.SubjectClaims.Federation != nil &&
				account.SubjectClaims.Federation.Id == federationID {
				return account, nil
			}
		}

		if listResp.NextPageToken == "" {
			break
		}

		nextPageToken = listResp.NextPageToken
	}

	return nil, fmt.Errorf("User '%s' from federation '%s' not found in organization '%s'", nameID, federationID, organizationID)
}
