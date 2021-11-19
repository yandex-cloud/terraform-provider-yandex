package yandex

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexOrganizationManagerSamlFederationDefaultTimeout = 1 * time.Minute

func resourceYandexOrganizationManagerSamlFederation() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexOrganizationManagerSamlFederationCreate,
		Read:   resourceYandexOrganizationManagerSamlFederationRead,
		Update: resourceYandexOrganizationManagerSamlFederationUpdate,
		Delete: resourceYandexOrganizationManagerSamlFederationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexOrganizationManagerSamlFederationDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexOrganizationManagerSamlFederationDefaultTimeout),
			Update: schema.DefaultTimeout(yandexOrganizationManagerSamlFederationDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexOrganizationManagerSamlFederationDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"organization_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"issuer": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sso_binding": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sso_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cookie_max_age": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validateParsableValue(parsePositiveDuration),
				DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
			},
			"auto_create_account_on_login": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"case_insensitive_name_ids": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"security_settings": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encrypted_assertions": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func getSamlFederationBindingTypes() string {
	var values []string
	for k := range saml.BindingType_value {
		values = append(values, k)
	}
	sort.Strings(values)

	return strings.Join(values, ",")
}

func getSamlFederationSSOBinding(d *schema.ResourceData) (saml.BindingType, error) {
	c, ok := d.GetOk("sso_binding")
	if ok {
		if bt, ok := saml.BindingType_value[c.(string)]; ok {
			return saml.BindingType(bt), nil
		}

		err := fmt.Errorf("invalid sso_binding field value, possible values: %v", getSamlFederationBindingTypes())
		return saml.BindingType_BINDING_TYPE_UNSPECIFIED, err
	}

	return saml.BindingType_BINDING_TYPE_UNSPECIFIED, nil
}

func getSamlFederationSecuritySettings(d *schema.ResourceData) *saml.FederationSecuritySettings {
	if _, ok := d.GetOk("security_settings"); !ok {
		return &saml.FederationSecuritySettings{}
	}
	return &saml.FederationSecuritySettings{
		EncryptedAssertions: d.Get("security_settings.0.encrypted_assertions").(bool),
	}
}

func flattenSamlFederationSecuritySettings(fss *saml.FederationSecuritySettings) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"encrypted_assertions": fss.GetEncryptedAssertions(),
		},
	}
}

func resourceYandexOrganizationManagerSamlFederationCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	organizationID, err := getOrganizationID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting organization ID while creating SAML Federation: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating SAML Federation: %s", err)
	}

	ssoBinding, err := getSamlFederationSSOBinding(d)
	if err != nil {
		return fmt.Errorf("Error getting SAML Federation SSO binding: %s", err)
	}

	cookieMaxAge, err := parseDuration(d.Get("cookie_max_age").(string))
	if err != nil {
		return err
	}

	req := saml.CreateFederationRequest{
		OrganizationId:           organizationID,
		Name:                     d.Get("name").(string),
		Labels:                   labels,
		Description:              d.Get("description").(string),
		Issuer:                   d.Get("issuer").(string),
		SsoBinding:               ssoBinding,
		SsoUrl:                   d.Get("sso_url").(string),
		CookieMaxAge:             cookieMaxAge,
		AutoCreateAccountOnLogin: d.Get("auto_create_account_on_login").(bool),
		CaseInsensitiveNameIds:   d.Get("case_insensitive_name_ids").(bool),
		SecuritySettings:         getSamlFederationSecuritySettings(d),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManagerSAML().Federation().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create SAML Federation: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get SAML Federation create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*saml.CreateFederationMetadata)
	if !ok {
		return fmt.Errorf("could not get SAML Federation ID from create operation metadata")
	}

	d.SetId(md.FederationId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create SAML Federation: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("SAML Federation creation failed: %s", err)
	}

	return resourceYandexOrganizationManagerSamlFederationRead(d, meta)
}

func resourceYandexOrganizationManagerSamlFederationRead(d *schema.ResourceData, meta interface{}) error {
	return flattenSamlFederation(d.Id(), d, meta)
}

func flattenSamlFederation(federationID string, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	federation, err := config.sdk.OrganizationManagerSAML().Federation().Get(context.Background(),
		&saml.GetFederationRequest{
			FederationId: federationID,
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("SAML Federation %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(federation.CreatedAt))
	d.Set("name", federation.Name)
	d.Set("organization_id", federation.OrganizationId)
	d.Set("description", federation.Description)
	d.Set("issuer", federation.Issuer)
	d.Set("sso_binding", federation.SsoBinding.String())
	d.Set("sso_url", federation.SsoUrl)
	d.Set("cookie_max_age", formatDuration(federation.CookieMaxAge))
	d.Set("auto_create_account_on_login", federation.AutoCreateAccountOnLogin)
	d.Set("case_insensitive_name_ids", federation.CaseInsensitiveNameIds)
	d.Set("security_settings", flattenSamlFederationSecuritySettings(federation.GetSecuritySettings()))

	return d.Set("labels", federation.Labels)
}

var updateSamlFederationFieldsMap = map[string]string{
	"name":                         "name",
	"description":                  "description",
	"issuer":                       "issuer",
	"sso_url":                      "sso_url",
	"sso_binding":                  "sso_binding",
	"cookie_max_age":               "cookie_max_age",
	"auto_create_account_on_login": "auto_create_account_on_login",
	"case_insensitive_name_ids":    "case_insensitive_name_ids",
	"security_settings.0.encrypted_assertions": "security_settings",
}

func resourceYandexOrganizationManagerSamlFederationUpdate(d *schema.ResourceData, meta interface{}) error {
	labelsProp, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	ssoBinding, err := getSamlFederationSSOBinding(d)
	if err != nil {
		return fmt.Errorf("Error getting SAML Federation SSO binding: %s", err)
	}

	cookieMaxAge, err := parseDuration(d.Get("cookie_max_age").(string))
	if err != nil {
		return err
	}

	req := &saml.UpdateFederationRequest{
		FederationId:             d.Id(),
		UpdateMask:               &field_mask.FieldMask{},
		Labels:                   labelsProp,
		Name:                     d.Get("name").(string),
		Description:              d.Get("description").(string),
		Issuer:                   d.Get("issuer").(string),
		SsoBinding:               ssoBinding,
		SsoUrl:                   d.Get("sso_url").(string),
		CookieMaxAge:             cookieMaxAge,
		AutoCreateAccountOnLogin: d.Get("auto_create_account_on_login").(bool),
		CaseInsensitiveNameIds:   d.Get("case_insensitive_name_ids").(bool),
		SecuritySettings:         getSamlFederationSecuritySettings(d),
	}

	var updatePath []string
	for field, path := range updateSamlFederationFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	if len(req.UpdateMask.Paths) == 0 {
		return fmt.Errorf("No fields were updated for SAML Federation %s", d.Id())
	}

	err = makeSamlFederationUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexOrganizationManagerSamlFederationRead(d, meta)
}

func resourceYandexOrganizationManagerSamlFederationDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting SAML Federation %q", d.Id())

	req := &saml.DeleteFederationRequest{
		FederationId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManagerSAML().Federation().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("SAML Federation %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting SAML Federation %q", d.Id())
	return nil
}

func makeSamlFederationUpdateRequest(req *saml.UpdateFederationRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManagerSAML().Federation().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update SAML Federation %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating SAML Federation %q: %s", d.Id(), err)
	}

	return nil
}
