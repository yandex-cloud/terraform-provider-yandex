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
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	samlsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/saml"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
)

const yandexOrganizationManagerSamlFederationUserDefaultTimeout = 1 * time.Minute

func resourceYandexOrganizationManagerSamlFederationUserAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of a single SAML Federation user account within an existing Yandex Cloud Organization.. For more information, see [the official documentation](https://yandex.cloud/docs/organization/operations/federations/integration-common).\n\n~> If terraform user has sufficient access and user specified in data source does not exist, it will be created. This behaviour will be **deprecated** in future releases. Use resource `yandex_organizationmanager_saml_federation_user_account` to manage account lifecycle.\n",

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
				Description:  "ID of a SAML Federation.",
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name_id": {
				Type:         schema.TypeString,
				Description:  "Name ID of the SAML federated user.",
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
	client := iamsdk.NewUserAccountClient(config.SDK)

	userAccount, err := client.Get(context, req)
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
	client := samlsdk.NewFederationClient(config.SDK)

	federationID, nameID := d.Get("federation_id").(string), d.Get("name_id").(string)
	req := &saml.AddFederatedUserAccountsRequest{
		FederationId: federationID,
		NameIds:      []string{nameID},
	}

	op, err := client.AddUserAccounts(config.Context(), req)
	if err != nil {
		return diag.Errorf("error on add user '%s' operation creation  into federation '%s': %s", nameID, federationID, err)
	}

	_, err = op.Wait(context)
	if err != nil {
		return diag.Errorf("error on add user '%s' operation wait into federation '%s': %s", nameID, federationID, err)
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
	client := organizationmanagersdk.NewUserClient(config.SDK)

	federationID, nameID := d.Get("federation_id").(string), d.Get("name_id").(string)

	federation, err := getSamlFederation(context, config, federationID)
	if err != nil {
		return diag.Errorf("error deleting saml user '%s': %s", nameID, err)
	}

	organizationID := federation.OrganizationId
	req := &organizationmanager.DeleteMembershipRequest{
		OrganizationId: organizationID,
		SubjectId:      d.Id(),
	}

	op, err := client.DeleteMembership(context, req)
	if err != nil {
		return diag.Errorf("error on delete saml user '%s' operation creation: %s", nameID, err)
	}

	_, err = op.Wait(context)
	if err != nil {
		return diag.Errorf("error on delete saml user '%s' operation wait: %s", nameID, err)
	}

	d.SetId("")

	return nil
}

func getSamlFederation(context context.Context, config *Config, federationID string) (*saml.Federation, error) {
	getFederationReq := &saml.GetFederationRequest{
		FederationId: federationID,
	}

	client := samlsdk.NewFederationClient(config.SDK)

	federation, err := client.Get(context, getFederationReq)
	if err != nil {
		return nil, fmt.Errorf("error on reading federation '%s': %s", federationID, err)
	}

	return federation, nil
}

func getSamlUserAccount(context context.Context, config *Config, federationID, nameID string) (*organizationmanager.ListMembersResponse_OrganizationUser, error) {
	client := organizationmanagersdk.NewUserClient(config.SDK)

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

		listResp, err := client.ListMembers(context, req)
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
