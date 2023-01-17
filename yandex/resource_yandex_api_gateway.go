package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexApiGatewayDefaultTimeout = 5 * time.Minute

func resourceYandexApiGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexApiGatewayCreate,
		Read:   resourceYandexApiGatewayRead,
		Update: resourceYandexApiGatewayUpdate,
		Delete: resourceYandexApiGatewayDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexApiGatewayDefaultTimeout),
			Update: schema.DefaultTimeout(yandexApiGatewayDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexApiGatewayDefaultTimeout),
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

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"spec": {
				Type:     schema.TypeString,
				Required: true,
			},

			"user_domains": {
				Type:       schema.TypeSet,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        schema.HashString,
				Deprecated: fieldDeprecatedForAnother("user_domains", "custom_domains"),
			},

			"custom_domains": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"certificate_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"log_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexApiGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	req, err := getCreateApiGatewayRequest(d, config)
	if err != nil {
		return err
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().APIGateway().ApiGateway().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud API Gateway: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud API Gateway: %s", err)
	}

	md, ok := protoMetadata.(*apigateway.CreateApiGatewayMetadata)
	if !ok {
		return fmt.Errorf("Could not get Yandex Cloud API Gateway ID from create operation metadata")
	}

	d.SetId(md.ApiGatewayId)
	d.Set("spec", d.Get("spec").(string))

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud API Gateway: %s", err)
	}

	// Attach custom domains
	customDomains, err := expandCustomDomains(d.Get("custom_domains"))
	if err != nil {
		return fmt.Errorf("Unable to construct custom_domains value: %s", err)
	}

	for _, customDomain := range customDomains {
		if err = attachDomain(ctx, config, md.ApiGatewayId, customDomain.Domain, customDomain.CertificateId); err != nil {
			return err
		}
	}

	return resourceYandexApiGatewayRead(d, meta)
}

func getCreateApiGatewayRequest(d *schema.ResourceData, config *Config) (*apigateway.CreateApiGatewayRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating Yandex Cloud API Gateway: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating Yandex Cloud API Gateway: %s", err)
	}

	req := &apigateway.CreateApiGatewayRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		Spec: &apigateway.CreateApiGatewayRequest_OpenapiSpec{
			OpenapiSpec: d.Get("spec").(string),
		},
	}

	return req, err
}

func resourceYandexApiGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while updating Yandex Cloud API Gateway: %s", err)
	}

	d.Partial(true)

	var updatePaths []string

	if d.HasChange("spec") {
		updatePaths = append(updatePaths, "openapi_spec")
	}

	if d.HasChange("name") {
		updatePaths = append(updatePaths, "name")
	}

	if d.HasChange("description") {
		updatePaths = append(updatePaths, "description")
	}

	if d.HasChange("labels") {
		updatePaths = append(updatePaths, "labels")
	}

	if len(updatePaths) != 0 {
		req := apigateway.UpdateApiGatewayRequest{
			ApiGatewayId: d.Id(),
			Name:         d.Get("name").(string),
			Description:  d.Get("description").(string),
			Labels:       labels,
			UpdateMask:   &field_mask.FieldMask{Paths: updatePaths},
			Spec: &apigateway.UpdateApiGatewayRequest_OpenapiSpec{
				OpenapiSpec: d.Get("spec").(string),
			},
		}

		op, err := config.sdk.Serverless().APIGateway().ApiGateway().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to update Yandex Cloud API Gateway: %s", err)
		}

	}

	if d.HasChanges("custom_domains") {
		oldVal, newVal := d.GetChange("custom_domains")

		oldDomains, err := expandCustomDomains(oldVal)
		if err != nil {
			return fmt.Errorf("Unable to construct previous custom_domains value: %s", err)
		}

		newDomains, err := expandCustomDomains(newVal)
		if err != nil {
			return fmt.Errorf("Unable to construct new custom_domains value: %s", err)
		}

		// Remove domains which are absent in new value
		for _, domain := range oldDomains {
			found := false

			for _, newDomain := range newDomains {
				if newDomain.DomainId == domain.DomainId {
					found = true
				}
			}

			if !found {
				if err = removeDomain(ctx, config, d.Id(), domain.DomainId); err != nil {
					return err
				}
			}
		}

		// Add new domains
		for _, domain := range newDomains {
			// Consider domains without ID as new ones
			if domain.DomainId == "" {
				if err = attachDomain(ctx, config, d.Id(), domain.Domain, domain.CertificateId); err != nil {
					return err
				}
			}
		}
	}

	d.Partial(false)

	return resourceYandexApiGatewayRead(d, meta)
}

func resourceYandexApiGatewayRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := apigateway.GetApiGatewayRequest{
		ApiGatewayId: d.Id(),
	}

	apiGateway, err := config.sdk.Serverless().APIGateway().ApiGateway().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud API Gateway %q", d.Id()))
	}

	return flattenYandexApiGateway(d, apiGateway)
}

func resourceYandexApiGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := apigateway.DeleteApiGatewayRequest{
		ApiGatewayId: d.Id(),
	}

	op, err := config.sdk.Serverless().APIGateway().ApiGateway().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud API Gateway %q", d.Id()))
	}

	return nil
}

func flattenYandexApiGateway(d *schema.ResourceData, apiGateway *apigateway.ApiGateway) error {
	d.Set("name", apiGateway.Name)
	d.Set("folder_id", apiGateway.FolderId)
	d.Set("description", apiGateway.Description)
	d.Set("created_at", getTimestamp(apiGateway.CreatedAt))
	d.Set("domain", apiGateway.Domain)
	d.Set("status", strings.ToLower(apiGateway.Status.String()))
	d.Set("log_group_id", apiGateway.LogGroupId)

	if err := d.Set("custom_domains", flattenCustomDomains(apiGateway.AttachedDomains)); err != nil {
		return fmt.Errorf("Unable to set custom_domains: %s, %#v", err)
	}

	domains := make([]string, len(apiGateway.AttachedDomains))
	for i, domain := range apiGateway.AttachedDomains {
		domains[i] = domain.DomainId
	}
	d.Set("user_domains", convertStringArrToInterface(domains))

	return d.Set("labels", apiGateway.Labels)
}

func attachDomain(ctx context.Context, config *Config, apigwID string, domain string, certificateId string) error {
	attachDomainRequest := &apigateway.AddDomainRequest{
		ApiGatewayId:  apigwID,
		DomainName:    domain,
		CertificateId: certificateId,
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().APIGateway().ApiGateway().AddDomain(ctx, attachDomainRequest))
	if err != nil {
		return fmt.Errorf("Error while requesting API to attach custom domain to Yandex Cloud API Gateway: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to attach custom domain to Yandex Cloud API Gateway: %s", err)
	}

	return nil
}

func removeDomain(ctx context.Context, config *Config, apigwID string, domainId string) error {
	removeDomainRequest := &apigateway.RemoveDomainRequest{
		ApiGatewayId: apigwID,
		DomainId:     domainId,
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().APIGateway().ApiGateway().RemoveDomain(ctx, removeDomainRequest))
	if err != nil {
		return fmt.Errorf("Error while requesting API to remove custom domain from Yandex Cloud API Gateway: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to remove custom domain from Yandex Cloud API Gateway: %s", err)
	}

	return nil
}
