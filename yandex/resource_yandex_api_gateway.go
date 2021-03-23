package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"io/ioutil"
	"strings"
	"time"
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
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"yaml_filename": {
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

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud API Gateway: %s", err)
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

	specContent, err := loadSpec(d.Get("spec.0.yaml_filename").(string))
	if err != nil {
		return nil, fmt.Errorf("Error getting spec while creating Yandex Cloud API Gateway: %s", err)
	}

	req := &apigateway.CreateApiGatewayRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		Spec: &apigateway.CreateApiGatewayRequest_OpenapiSpec{
			OpenapiSpec: *specContent,
		},
	}
	return req, err
}

func resourceYandexApiGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while updating Yandex Cloud API Gateway: %s", err)
	}

	d.Partial(true)

	var updatePaths []string
	var partialPaths []string
	if d.HasChange("name") {
		updatePaths = append(updatePaths, "name")
		partialPaths = append(partialPaths, "name")
	}

	if d.HasChange("description") {
		updatePaths = append(updatePaths, "description")
		partialPaths = append(partialPaths, "description")
	}

	if d.HasChange("labels") {
		updatePaths = append(updatePaths, "labels")
		partialPaths = append(partialPaths, "labels")
	}

	if len(updatePaths) != 0 {
		specContent, err := loadSpec(d.Get("spec.0.yaml_filename").(string))
		if err != nil {
			return err
		}
		req := apigateway.UpdateApiGatewayRequest{
			ApiGatewayId: d.Id(),
			Name:         d.Get("name").(string),
			Description:  d.Get("description").(string),
			Labels:       labels,
			UpdateMask:   &field_mask.FieldMask{Paths: updatePaths},
			Spec: &apigateway.UpdateApiGatewayRequest_OpenapiSpec{
				OpenapiSpec: *specContent,
			},
		}

		op, err := config.sdk.Serverless().APIGateway().ApiGateway().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to update Yandex Cloud API Gateway: %s", err)
		}

		for _, v := range partialPaths {
			d.SetPartial(v)
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

func loadSpec(filename string) (*string, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	specString := string(fileBytes)
	return &specString, nil
}

func flattenYandexApiGateway(d *schema.ResourceData, apiGateway *apigateway.ApiGateway) error {
	createdAt, err := getTimestamp(apiGateway.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("name", apiGateway.Name)
	d.Set("folder_id", apiGateway.FolderId)
	d.Set("description", apiGateway.Description)
	d.Set("created_at", createdAt)
	d.Set("domain", apiGateway.Domain)
	d.Set("status", strings.ToLower(apiGateway.Status.String()))
	d.Set("log_group_id", apiGateway.LogGroupId)
	return d.Set("labels", apiGateway.Labels)
}
