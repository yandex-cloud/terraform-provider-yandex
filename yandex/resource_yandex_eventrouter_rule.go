package yandex

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexEventrouterRuleDefaultTimeout = 10 * time.Minute

	maxEventRouterRuleTargetsCount = 5

	eventrouterTargetTypeYds                       = "yds"
	eventrouterTargetTypeYmq                       = "ymq"
	eventrouterTargetTypeFunction                  = "function"
	eventrouterTargetTypeContainer                 = "container"
	eventrouterTargetTypeGatewayWebsocketBroadcast = "gateway_websocket_broadcast"
	eventrouterTargetTypeLogging                   = "logging"
	eventrouterTargetTypeWorkflow                  = "workflow"

	eventrouterFilterTypeJq = "jq_filter"
)

var yandexEventrouterTargetTypesList = []string{
	eventrouterTargetTypeYds,
	eventrouterTargetTypeYmq,
	eventrouterTargetTypeFunction,
	eventrouterTargetTypeContainer,
	eventrouterTargetTypeGatewayWebsocketBroadcast,
	eventrouterTargetTypeLogging,
	eventrouterTargetTypeWorkflow,
}

var yandexEventrouterTargetBatchSettingsResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"max_count": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "Maximum batch size: rule will send a batch if number of events exceeds this value",
		},

		"max_bytes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "Maximum batch size: rule will send a batch if total size of events exceeds this value",
		},

		"cutoff": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Maximum batch size: rule will send a batch if its lifetime exceeds this value",
		},
	},
}

func resourceYandexEventrouterRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexEventrouterRuleCreate,
		ReadContext:   resourceYandexEventrouterRuleRead,
		UpdateContext: resourceYandexEventrouterRuleUpdate,
		DeleteContext: resourceYandexEventrouterRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexEventrouterRuleDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexEventrouterRuleDefaultTimeout),
			Update: schema.DefaultTimeout(yandexEventrouterRuleDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexEventrouterRuleDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the rule",
			},

			"bus_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the bus that the rule belongs to",
			},

			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the folder that the rule resides in",
			},

			"cloud_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud that the rule resides in",
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the rule",
			},

			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "Rule labels",
			},

			eventrouterFilterTypeJq: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "JQ filter for matching events",
			},

			eventrouterTargetTypeYds: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "YdsTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Stream database",
						},

						"stream_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Full stream name, like /ru-central1/aoegtvhtp8ob********/cc8004q4lbo6********/test",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Service account, which has write permission on the stream",
						},
					},
				},
			},

			eventrouterTargetTypeYmq: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "YmqTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_arn": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Queue ARN. Example: yrn:yc:ymq:ru-central1:aoe***:test",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Service account which has write access to the queue",
						},
					},
				},
			},

			eventrouterTargetTypeFunction: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "FunctionTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Function ID",
						},

						"function_tag": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Function tag",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Service account which has call permission on the function",
						},

						"batch_settings": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem:        yandexEventrouterTargetBatchSettingsResource,
							Description: "Batch settings",
						},
					},
				},
			},

			eventrouterTargetTypeContainer: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "ContainerTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Container ID",
						},

						"container_revision_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Container revision ID",
						},

						"path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Endpoint HTTP path to invoke",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Service account which should be used to call a container",
						},

						"batch_settings": {
							Type:        schema.TypeList,
							Computed:    true,
							Optional:    true,
							MaxItems:    1,
							Elem:        yandexEventrouterTargetBatchSettingsResource,
							Description: "Batch settings",
						},
					},
				},
			},

			eventrouterTargetTypeGatewayWebsocketBroadcast: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "GatewayWebsocketBroadcastTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Gateway ID",
						},

						"path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Path",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Service account which has permission for writing to websockets",
						},

						"batch_settings": {
							Type:        schema.TypeList,
							Computed:    true,
							Optional:    true,
							MaxItems:    1,
							Elem:        yandexEventrouterTargetBatchSettingsResource,
							Description: "Batch settings",
						},
					},
				},
			},

			eventrouterTargetTypeLogging: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "LoggingTarget. Includes either log_group_id or folder_id",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Log group ID",
						},
						"folder_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Folder ID",
						},
						"service_account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Service account which has permission for writing logs",
						},
					},
				},
			},

			eventrouterTargetTypeWorkflow: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "WorkflowTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"workflow_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Workflow ID",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Service account which should be used to start workflow",
						},

						"batch_settings": {
							Type:        schema.TypeList,
							Computed:    true,
							Optional:    true,
							MaxItems:    1,
							Elem:        yandexEventrouterTargetBatchSettingsResource,
							Description: "Batch settings",
						},
					},
				},
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Deletion protection",
			},
		},
	}
}

func flattenYandexEventrouterRule(
	d *schema.ResourceData,
	rule *eventrouter.Rule,
) {
	d.Set("name", rule.Name)
	d.Set("bus_id", rule.BusId)
	d.Set("folder_id", rule.FolderId)
	d.Set("cloud_id", rule.CloudId)
	d.Set("created_at", getTimestamp(rule.CreatedAt))
	d.Set("description", rule.Description)
	d.Set("labels", rule.Labels)
	if rule.Filter != nil {
		if jqFilter, ok := rule.Filter.Condition.(*eventrouter.Filter_JqFilter); ok {
			d.Set("jq_filter", jqFilter.JqFilter)
		}
	}
	flattenYandexEventrouterTargets(d, rule.Targets)
	d.Set("deletion_protection", rule.DeletionProtection)
}

func eventrouterTargetsAppend(targetsByType map[string][]interface{}, targetType string, targetEntry interface{}) {
	targetsByType[targetType] = append(targetsByType[targetType], targetEntry)
}

func flattenYandexEventrouterTargets(
	d *schema.ResourceData,
	targets []*eventrouter.Target,
) {
	targetsByType := make(map[string][]interface{})
	for _, target := range targets {
		switch t := target.Target.(type) {
		case *eventrouter.Target_Yds:
			yds := t.Yds
			targetEntry := map[string]interface{}{
				"database":           yds.Database,
				"stream_name":        yds.StreamName,
				"service_account_id": yds.ServiceAccountId,
			}
			eventrouterTargetsAppend(targetsByType, eventrouterTargetTypeYds, targetEntry)
		case *eventrouter.Target_Ymq:
			ymq := t.Ymq
			targetEntry := map[string]interface{}{
				"queue_arn":          ymq.QueueArn,
				"service_account_id": ymq.ServiceAccountId,
			}
			eventrouterTargetsAppend(targetsByType, eventrouterTargetTypeYmq, targetEntry)
		case *eventrouter.Target_Function:
			function := t.Function
			targetEntry := map[string]interface{}{
				"function_id":        function.FunctionId,
				"function_tag":       function.FunctionTag,
				"service_account_id": function.ServiceAccountId,
				"batch_settings":     flattenYandexEventrouterTargetBatchSettings(function.BatchSettings),
			}
			eventrouterTargetsAppend(targetsByType, eventrouterTargetTypeFunction, targetEntry)
		case *eventrouter.Target_Container:
			container := t.Container
			targetEntry := map[string]interface{}{
				"container_id":          container.ContainerId,
				"container_revision_id": container.ContainerRevisionId,
				"path":                  container.Path,
				"service_account_id":    container.ServiceAccountId,
				"batch_settings":        flattenYandexEventrouterTargetBatchSettings(container.BatchSettings),
			}
			eventrouterTargetsAppend(targetsByType, eventrouterTargetTypeContainer, targetEntry)
		case *eventrouter.Target_GatewayWsBroadcast:
			gatewayWsBroadcast := t.GatewayWsBroadcast
			targetEntry := map[string]interface{}{
				"gateway_id":         gatewayWsBroadcast.GatewayId,
				"path":               gatewayWsBroadcast.Path,
				"service_account_id": gatewayWsBroadcast.ServiceAccountId,
				"batch_settings":     flattenYandexEventrouterTargetBatchSettings(gatewayWsBroadcast.BatchSettings),
			}
			eventrouterTargetsAppend(targetsByType, eventrouterTargetTypeGatewayWebsocketBroadcast, targetEntry)
		case *eventrouter.Target_Logging:
			var targetEntry interface{}
			logging := t.Logging
			switch destination := t.Logging.Destination.(type) {
			case *eventrouter.LoggingTarget_FolderId:
				targetEntry = map[string]interface{}{
					"folder_id":          destination.FolderId,
					"service_account_id": logging.ServiceAccountId,
				}
			case *eventrouter.LoggingTarget_LogGroupId:
				targetEntry = map[string]interface{}{
					"log_group_id":       destination.LogGroupId,
					"service_account_id": logging.ServiceAccountId,
				}
			}
			eventrouterTargetsAppend(targetsByType, eventrouterTargetTypeLogging, targetEntry)
		case *eventrouter.Target_Workflow:
			workflow := t.Workflow
			targetEntry := map[string]interface{}{
				"workflow_id":        workflow.WorkflowId,
				"service_account_id": workflow.ServiceAccountId,
				"batch_settings":     flattenYandexEventrouterTargetBatchSettings(workflow.BatchSettings),
			}
			eventrouterTargetsAppend(targetsByType, eventrouterTargetTypeWorkflow, targetEntry)
		}
		for targetType, targetEntries := range targetsByType {
			d.Set(targetType, targetEntries)
		}
	}
}

func flattenYandexEventrouterTargetBatchSettings(batchSettings *eventrouter.BatchSettings) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"max_count": batchSettings.MaxCount,
			"max_bytes": batchSettings.MaxBytes,
			"cutoff":    formatDuration(batchSettings.Cutoff),
		},
	}
}

func constructYandexEventrouterTargetBatchSettings(schema interface{}) (*eventrouter.BatchSettings, error) {
	bsList := schema.([]interface{})
	if len(bsList) == 0 {
		return nil, nil
	}
	bs := bsList[0].(map[string]interface{})
	cutoff, err := parseDuration(bs["cutoff"].(string))
	if err != nil {
		return nil, err
	}
	return &eventrouter.BatchSettings{
		MaxCount: int64(bs["max_count"].(int)),
		MaxBytes: int64(bs["max_bytes"].(int)),
		Cutoff:   cutoff,
	}, nil
}

func constructYandexEventrouterTargets(d *schema.ResourceData) ([]*eventrouter.Target, error) {
	var targets []*eventrouter.Target
	for _, value := range d.Get(eventrouterTargetTypeYds).([]interface{}) {
		yds := value.(map[string]interface{})
		targets = append(targets, &eventrouter.Target{
			Target: &eventrouter.Target_Yds{
				Yds: &eventrouter.YdsTarget{
					Database:         yds["database"].(string),
					StreamName:       yds["stream_name"].(string),
					ServiceAccountId: yds["service_account_id"].(string),
				},
			},
		})
	}
	for _, value := range d.Get(eventrouterTargetTypeYmq).([]interface{}) {
		ymq := value.(map[string]interface{})
		targets = append(targets, &eventrouter.Target{
			Target: &eventrouter.Target_Ymq{
				Ymq: &eventrouter.YmqTarget{
					QueueArn:         ymq["queue_arn"].(string),
					ServiceAccountId: ymq["service_account_id"].(string),
				},
			},
		})
	}
	for _, value := range d.Get(eventrouterTargetTypeFunction).([]interface{}) {
		function := value.(map[string]interface{})
		bs, err := constructYandexEventrouterTargetBatchSettings(function["batch_settings"])
		if err != nil {
			return nil, err
		}
		targets = append(targets, &eventrouter.Target{
			Target: &eventrouter.Target_Function{
				Function: &eventrouter.FunctionTarget{
					FunctionId:       function["function_id"].(string),
					FunctionTag:      function["function_tag"].(string),
					ServiceAccountId: function["service_account_id"].(string),
					BatchSettings:    bs,
				},
			},
		})
	}
	for _, value := range d.Get(eventrouterTargetTypeContainer).([]interface{}) {
		container := value.(map[string]interface{})
		bs, err := constructYandexEventrouterTargetBatchSettings(container["batch_settings"])
		if err != nil {
			return nil, err
		}
		targets = append(targets, &eventrouter.Target{
			Target: &eventrouter.Target_Container{
				Container: &eventrouter.ContainerTarget{
					ContainerId:         container["container_id"].(string),
					ContainerRevisionId: container["container_revision_id"].(string),
					Path:                container["path"].(string),
					ServiceAccountId:    container["service_account_id"].(string),
					BatchSettings:       bs,
				},
			},
		})
	}
	for _, value := range d.Get(eventrouterTargetTypeGatewayWebsocketBroadcast).([]interface{}) {
		gatewayWsBroadcast := value.(map[string]interface{})
		bs, err := constructYandexEventrouterTargetBatchSettings(gatewayWsBroadcast["batch_settings"])
		if err != nil {
			return nil, err
		}
		targets = append(targets, &eventrouter.Target{
			Target: &eventrouter.Target_GatewayWsBroadcast{
				GatewayWsBroadcast: &eventrouter.GatewayWebsocketBroadcastTarget{
					GatewayId:        gatewayWsBroadcast["gateway_id"].(string),
					Path:             gatewayWsBroadcast["path"].(string),
					ServiceAccountId: gatewayWsBroadcast["service_account_id"].(string),
					BatchSettings:    bs,
				},
			},
		})
	}
	for _, value := range d.Get(eventrouterTargetTypeLogging).([]interface{}) {
		logging := value.(map[string]interface{})
		var destination eventrouter.LoggingTarget_Destination
		logGroupId := logging["log_group_id"]
		folderId := logging["folder_id"]

		if (logGroupId == "") == (folderId == "") {
			return nil, errors.New("For Event Router logging target exactly one of log_group_id and folder_id must be specified")
		}

		if logGroupId != "" {
			destination = &eventrouter.LoggingTarget_LogGroupId{
				LogGroupId: logGroupId.(string),
			}
		} else { // use folderId
			destination = &eventrouter.LoggingTarget_FolderId{
				FolderId: folderId.(string),
			}
		}
		targets = append(targets, &eventrouter.Target{
			Target: &eventrouter.Target_Logging{
				Logging: &eventrouter.LoggingTarget{
					Destination:      destination,
					ServiceAccountId: logging["service_account_id"].(string),
				},
			},
		})
	}
	for _, value := range d.Get(eventrouterTargetTypeWorkflow).([]interface{}) {
		workflow := value.(map[string]interface{})
		bs, err := constructYandexEventrouterTargetBatchSettings(workflow["batch_settings"])
		if err != nil {
			return nil, err
		}
		targets = append(targets, &eventrouter.Target{
			Target: &eventrouter.Target_Workflow{
				Workflow: &eventrouter.WorkflowTarget{
					WorkflowId:       workflow["workflow_id"].(string),
					ServiceAccountId: workflow["service_account_id"].(string),
					BatchSettings:    bs,
				},
			},
		})
	}

	if len(targets) == 0 {
		return nil, errors.New("No targets are specified for Event Router rule")
	}
	if total := len(targets); total > maxEventRouterRuleTargetsCount {
		return nil, fmt.Errorf("Too many targets for Event Router rule, expected [1, %d] but got %d",
			maxEventRouterRuleTargetsCount, total)
	}

	return targets, nil
}

func constructYandexEventrouterFilter(d *schema.ResourceData) (*eventrouter.Filter, error) {
	if value, ok := d.GetOk(eventrouterFilterTypeJq); ok {
		return &eventrouter.Filter{
			Condition: &eventrouter.Filter_JqFilter{
				JqFilter: value.(string),
			},
		}, nil
	}
	return nil, nil
}

func resourceYandexEventrouterRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while creating Event Router rule: %s", err)
	}

	busId := d.Get("bus_id").(string)

	filter, err := constructYandexEventrouterFilter(d)
	if err != nil {
		return diag.Errorf("Error constructing Event Router rule filter: %s", err)
	}

	targets, err := constructYandexEventrouterTargets(d)
	if err != nil {
		return diag.Errorf("Error constructing Event Router rule targets: %s", err)
	}

	req := eventrouter.CreateRuleRequest{
		BusId:              busId,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		Filter:             filter,
		Targets:            targets,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Eventrouter().Rule().Create(ctx, &req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router rule: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router rule: %s", err)
	}

	md, ok := protoMetadata.(*eventrouter.CreateRuleMetadata)
	if !ok {
		return diag.Errorf("Could not get Event Router rule ID from create operation metadata")
	}

	d.SetId(md.RuleId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router rule: %s", err)
	}

	return resourceYandexEventrouterRuleRead(ctx, d, meta)
}

func resourceYandexEventrouterRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := eventrouter.GetRuleRequest{
		RuleId: d.Id(),
	}

	rule, err := config.sdk.Serverless().Eventrouter().Rule().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router connector %q", d.Id())))
	}

	flattenYandexEventrouterRule(d, rule)
	return nil
}

func yandexEventrouterRuleTargetsChanged(d *schema.ResourceData) bool {
	for _, target := range yandexEventrouterTargetTypesList {
		if d.HasChange(target) {
			return true
		}
	}
	return false
}

func resourceYandexEventrouterRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while updating Event Router rule: %s", err)
	}

	var updatePaths []string
	if d.HasChange("name") {
		updatePaths = append(updatePaths, "name")
	}

	if d.HasChange("description") {
		updatePaths = append(updatePaths, "description")
	}

	if d.HasChange("labels") {
		updatePaths = append(updatePaths, "labels")
	}

	if d.HasChange("filter") {
		updatePaths = append(updatePaths, "filter")
	}

	if yandexEventrouterRuleTargetsChanged(d) {
		updatePaths = append(updatePaths, "targets")
	}

	if d.HasChange("deletion_protection") {
		updatePaths = append(updatePaths, "deletion_protection")
	}

	filter, err := constructYandexEventrouterFilter(d)
	if err != nil {
		return diag.Errorf("Error constructing Event Router rule filter: %s", err)
	}

	targets, err := constructYandexEventrouterTargets(d)
	if err != nil {
		return diag.Errorf("Error constructing Event Router rule filter: %s", err)
	}

	if len(updatePaths) != 0 {
		req := eventrouter.UpdateRuleRequest{
			RuleId:             d.Id(),
			UpdateMask:         &field_mask.FieldMask{Paths: updatePaths},
			Name:               d.Get("name").(string),
			Description:        d.Get("description").(string),
			Labels:             labels,
			Filter:             filter,
			Targets:            targets,
			DeletionProtection: d.Get("deletion_protection").(bool),
		}

		op, err := config.sdk.Serverless().Eventrouter().Rule().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return diag.Errorf("Error while requesting API to update Event Router rule: %s", err)
		}
	}

	return resourceYandexEventrouterRuleRead(ctx, d, meta)
}

func resourceYandexEventrouterRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := eventrouter.DeleteRuleRequest{
		RuleId: d.Id(),
	}

	op, err := config.sdk.Serverless().Eventrouter().Rule().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router rule %q", d.Id())))
	}

	return nil
}
