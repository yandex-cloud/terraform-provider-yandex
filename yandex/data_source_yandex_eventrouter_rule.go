package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

var yandexEventrouterTargetBatchSettingsDataSource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"max_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Maximum batch size: rule will send a batch if number of events exceeds this value",
		},

		"max_bytes": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Maximum batch size: rule will send a batch if total size of events exceeds this value",
		},

		"cutoff": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Maximum batch size: rule will send a batch if its lifetime exceeds this value",
		},
	},
}

func dataSourceYandexEventrouterRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexEventrouterRuleRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the rule",
			},

			"rule_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the bus that the rule belongs to",
			},

			"bus_id": {
				Type:        schema.TypeString,
				Computed:    true,
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
				Computed:    true,
				Description: "Description of the rule",
			},

			"labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "Rule labels",
			},

			"jq_filter": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JQ filter for matching events",
			},

			eventrouterTargetTypeYds: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "YdsTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Stream database",
						},

						"stream_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Full stream name, like /ru-central1/aoegtvhtp8ob********/cc8004q4lbo6********/test",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account, which has write permission on the stream",
						},
					},
				},
			},

			eventrouterTargetTypeYmq: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "YmqTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_arn": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Queue ARN. Example: yrn:yc:ymq:ru-central1:aoe***:test",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account which has write access to the queue",
						},
					},
				},
			},

			eventrouterTargetTypeFunction: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "FunctionTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Function ID",
						},

						"function_tag": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Function tag",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account which has call permission on the function",
						},

						"batch_settings": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        yandexEventrouterTargetBatchSettingsDataSource,
							Description: "Batch settings",
						},
					},
				},
			},

			eventrouterTargetTypeContainer: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "ContainerTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Container ID",
						},

						"container_revision_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Container revision ID",
						},

						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Endpoint HTTP path to invoke",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account which should be used to call a container",
						},

						"batch_settings": {
							Type:        schema.TypeString,
							Computed:    true,
							Elem:        yandexEventrouterTargetBatchSettingsDataSource,
							Description: "Batch settings",
						},
					},
				},
			},

			eventrouterTargetTypeGatewayWebsocketBroadcast: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "GatewayWebsocketBroadcastTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Gateway ID",
						},

						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Path",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account which has permission for writing to websockets",
						},

						"batch_settings": {
							Type:        schema.TypeString,
							Computed:    true,
							Elem:        yandexEventrouterTargetBatchSettingsDataSource,
							Description: "Batch settings",
						},
					},
				},
			},

			eventrouterTargetTypeLogging: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "LoggingTarget. Includes either log_group_id or folder_id",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Log group ID",
						},

						"folder_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Folder ID",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account which has permission for writing logs",
						},
					},
				},
			},

			eventrouterTargetTypeWorkflow: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "WorkflowTarget",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"workflow_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Workflow ID",
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account which should be used to start workflow",
						},

						"batch_settings": {
							Type:        schema.TypeString,
							Computed:    true,
							Elem:        yandexEventrouterTargetBatchSettingsDataSource,
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

func dataSourceYandexEventrouterRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "rule_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	ruleId := d.Get("rule_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		ruleId, err = resolveObjectID(ctx, config, d, sdkresolvers.EventrouterRuleResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source Event Router rule by name: %v", err)
		}
	}

	req := eventrouter.GetRuleRequest{
		RuleId: ruleId,
	}

	rule, err := config.sdk.Serverless().Eventrouter().Rule().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router rule %q", d.Id())))
	}

	d.SetId(rule.Id)
	d.Set("rule_id", rule.Id)
	flattenYandexEventrouterRule(d, rule)
	return nil
}
