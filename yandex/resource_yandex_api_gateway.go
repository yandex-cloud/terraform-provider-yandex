package yandex

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/durationpb"
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
						"fqdn": {
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

			"connectivity": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"canary": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 99),
						},
						"variables": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},

			"log_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"log_group_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"log_options.0.folder_id"},
						},
						"folder_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"log_options.0.log_group_id"},
						},
						"min_level": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"execution_timeout": {
				Type:     schema.TypeString,
				Optional: true,
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

	logOptions, err := expandApiGatewayLogOptions(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding log options while creating Yandex Cloud API Gateway: %s", err)
	}

	executionTimeout, err := expandApiGatewayExecutionTimeout(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding execution timeout while creating Yandex Cloud API Gateway: %s", err)
	}

	req := &apigateway.CreateApiGatewayRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		Spec: &apigateway.CreateApiGatewayRequest_OpenapiSpec{
			OpenapiSpec: d.Get("spec").(string),
		},
		Variables:        expandApiGatewayVariables(d),
		Canary:           expandApiGatewayCanary(d),
		LogOptions:       logOptions,
		ExecutionTimeout: executionTimeout,
	}

	if connectivity := expandApiGatewayConnectivity(d); connectivity != nil {
		req.Connectivity = connectivity
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

	if d.HasChange("connectivity") {
		updatePaths = append(updatePaths, "connectivity")
	}

	if d.HasChange("variables") {
		updatePaths = append(updatePaths, "variables")
	}

	if d.HasChange("canary") {
		updatePaths = append(updatePaths, "canary")
	}

	if d.HasChange("log_options") {
		updatePaths = append(updatePaths, "log_options")
	}

	if d.HasChange("execution_timeout") {
		updatePaths = append(updatePaths, "execution_timeout")
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
			Variables: expandApiGatewayVariables(d),
			Canary:    expandApiGatewayCanary(d),
		}

		if connectivity := expandApiGatewayConnectivity(d); connectivity != nil {
			req.Connectivity = connectivity
		}

		if logOptions, err := expandApiGatewayLogOptions(d); err != nil {
			return fmt.Errorf("Error expanding log options while updating Yandex Cloud API Gateway: %s", err)
		} else {
			req.LogOptions = logOptions
		}

		if executionTimeout, err := expandApiGatewayExecutionTimeout(d); err != nil {
			return fmt.Errorf("Error expanding execution timeout while updating Yandex Cloud API Gateway: %s", err)
		} else {
			req.ExecutionTimeout = executionTimeout
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

	return flattenYandexApiGateway(d, apiGateway, false)
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

func flattenYandexApiGateway(d *schema.ResourceData, apiGateway *apigateway.ApiGateway, allFields bool) error {
	d.Set("name", apiGateway.Name)
	d.Set("folder_id", apiGateway.FolderId)
	d.Set("description", apiGateway.Description)
	d.Set("created_at", getTimestamp(apiGateway.CreatedAt))
	d.Set("domain", apiGateway.Domain)
	d.Set("status", strings.ToLower(apiGateway.Status.String()))
	d.Set("log_group_id", apiGateway.LogGroupId)
	d.Set("labels", apiGateway.Labels)
	d.Set("custom_domains", flattenCustomDomains(apiGateway.AttachedDomains))

	if connectivity := flattenApiGatewayConnectivity(apiGateway.Connectivity); connectivity != nil {
		d.Set("connectivity", connectivity)
	}
	if variables := flattenApiGatewayVariables(apiGateway.Variables); variables != nil {
		d.Set("variables", variables)
	}
	if canary := flattenApiGatewayCanary(apiGateway.Canary); canary != nil {
		d.Set("canary", canary)
	}
	d.Set("log_options", flattenApiGatewayLogOptions(d, apiGateway.LogOptions, apiGateway.FolderId, allFields))
	if apiGateway.ExecutionTimeout != nil && apiGateway.ExecutionTimeout.Seconds != 0 {
		d.Set("execution_timeout", strconv.FormatInt(apiGateway.ExecutionTimeout.Seconds, 10))
	}
	domains := make([]string, len(apiGateway.AttachedDomains))
	for i, domain := range apiGateway.AttachedDomains {
		domains[i] = domain.DomainId
	}
	d.Set("user_domains", convertStringArrToInterface(domains))

	return nil
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

func expandApiGatewayConnectivity(d *schema.ResourceData) *apigateway.Connectivity {
	if id, ok := d.GetOk("connectivity.0.network_id"); ok {
		return &apigateway.Connectivity{NetworkId: id.(string)}
	}
	return nil
}

func flattenApiGatewayConnectivity(connectivity *apigateway.Connectivity) []interface{} {
	if connectivity == nil || connectivity.NetworkId == "" {
		return nil
	}
	return []interface{}{map[string]interface{}{"network_id": connectivity.NetworkId}}
}

func expandApiGatewayVariables(d *schema.ResourceData) map[string]*apigateway.VariableInput {
	v, ok := d.GetOk("variables")
	if !ok {
		return nil
	}
	return expandVariables(v)
}

func expandVariables(v interface{}) map[string]*apigateway.VariableInput {
	result := make(map[string]*apigateway.VariableInput)
	for key, value := range v.(map[string]interface{}) {
		var variable *apigateway.VariableInput
		if v, err := strconv.ParseInt(value.(string), 10, 64); err == nil {
			variable = &apigateway.VariableInput{VariableValue: &apigateway.VariableInput_IntValue{IntValue: v}}
		} else if v, err := strconv.ParseFloat(value.(string), 64); err == nil {
			variable = &apigateway.VariableInput{VariableValue: &apigateway.VariableInput_DoubleValue{DoubleValue: v}}
		} else if v, err := strconv.ParseBool(value.(string)); err == nil {
			variable = &apigateway.VariableInput{VariableValue: &apigateway.VariableInput_BoolValue{BoolValue: v}}
		} else {
			variable = &apigateway.VariableInput{VariableValue: &apigateway.VariableInput_StringValue{StringValue: value.(string)}}
		}
		result[key] = variable
	}
	return result
}

func flattenApiGatewayVariables(variables map[string]*apigateway.VariableInput) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range variables {
		var variable string
		switch value.VariableValue.(type) {
		case *apigateway.VariableInput_IntValue:
			variable = strconv.FormatInt(value.GetIntValue(), 10)
		case *apigateway.VariableInput_DoubleValue:
			variable = strconv.FormatFloat(value.GetDoubleValue(), 'f', -1, 64)
		case *apigateway.VariableInput_BoolValue:
			variable = strconv.FormatBool(value.GetBoolValue())
		default:
			variable = value.GetStringValue()
		}
		result[key] = variable
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func flattenApiGatewayCanary(canary *apigateway.Canary) []interface{} {
	if canary == nil || len(canary.Variables) == 0 {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"weight":    canary.Weight,
			"variables": flattenApiGatewayVariables(canary.Variables),
		},
	}
}

func expandApiGatewayCanary(d *schema.ResourceData) *apigateway.Canary {
	w, okW := d.GetOk("canary.0.weight")
	variables, okV := d.GetOk("canary.0.variables")
	weight, okI := w.(int)
	if !okW || !okV || !okI {
		return nil
	}
	return &apigateway.Canary{
		Weight:    int64(weight),
		Variables: expandVariables(variables),
	}
}

func expandApiGatewayLogOptions(d *schema.ResourceData) (*apigateway.LogOptions, error) {
	v, ok := d.GetOk("log_options.0")
	if !ok {
		return nil, nil
	}
	logOptionsMap := v.(map[string]interface{})
	if logOptionsMap["disabled"].(bool) {
		return &apigateway.LogOptions{
			Disabled: true,
		}, nil
	}
	logOptions := &apigateway.LogOptions{}
	if folderID, ok := logOptionsMap["folder_id"]; ok {
		logOptions.SetFolderId(folderID.(string))
	}
	if logGroupID, ok := logOptionsMap["log_group_id"]; ok {
		logOptions.SetLogGroupId(logGroupID.(string))
	}
	if level := logOptionsMap["min_level"]; len(level.(string)) > 0 {
		logLevel, ok := logging.LogLevel_Level_value[level.(string)]
		if !ok {
			return nil, fmt.Errorf("unknown log level: %s", level)
		}
		logOptions.MinLevel = logging.LogLevel_Level(logLevel)
	}
	return logOptions, nil
}

func flattenApiGatewayLogOptions(
	d *schema.ResourceData,
	logOptions *apigateway.LogOptions,
	apigatewayFolderID string,
	allFields bool,
) []interface{} {
	if logOptions == nil {
		return nil
	}
	res := make(map[string]interface{})
	if !allFields && logOptions.Disabled {
		res["disabled"] = true
		return []interface{}{res}
	}
	if allFields || len(d.Get("log_options.0.min_level").(string)) > 0 || logOptions.MinLevel != 0 {
		res["min_level"] = logging.LogLevel_Level_name[int32(logOptions.MinLevel)]
	}
	if logOptions.Destination != nil {
		switch destination := logOptions.Destination.(type) {
		case *apigateway.LogOptions_LogGroupId:
			res["log_group_id"] = destination.LogGroupId
		case *apigateway.LogOptions_FolderId:
			if allFields ||
				len(d.Get("log_options.0.folder_id").(string)) > 0 ||
				destination.FolderId != apigatewayFolderID {

				res["folder_id"] = destination.FolderId
			}
		}
	}
	if !allFields && len(d.Get("log_options").([]interface{})) <= 0 && len(res) <= 0 {
		return nil
	}
	res["disabled"] = logOptions.Disabled
	return []interface{}{res}
}

func expandApiGatewayExecutionTimeout(d *schema.ResourceData) (*durationpb.Duration, error) {
	strTimeout, ok := d.GetOk("execution_timeout")
	if !ok {
		return nil, nil
	}

	timeout, err := strconv.ParseInt(strTimeout.(string), 10, 64)
	if err != nil {
		return nil, err
	}

	return &durationpb.Duration{Seconds: timeout}, nil
}
