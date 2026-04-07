package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexFunctionTrigger() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Cloud Function Trigger. For more information about Yandex Cloud Functions, see [Yandex Cloud Functions](https://yandex.cloud/docs/functions/).\n\nThis data source is used to define [Yandex Cloud Functions Trigger](https://yandex.cloud/docs/functions/concepts/trigger) that can be used by other resources.\n\n~> Either `trigger_id` or `name` must be specified.\n",

		Read: dataSourceYandexFunctionTriggerRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"trigger_id": {
				Type:        schema.TypeString,
				Description: "Yandex Cloud Functions Trigger id used to define trigger.",
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			triggerTypeIoT: {
				Type:        schema.TypeList,
				Description: "[IoT](https://yandex.cloud/docs/functions/concepts/trigger/iot-core-trigger) settings definition for Yandex Cloud Functions Trigger, if present. Only one section `iot` or `message_queue`.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"registry_id": {
							Type:        schema.TypeString,
							Description: "IoT Registry ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"device_id": {
							Type:        schema.TypeString,
							Description: "IoT Device ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"topic": {
							Type:        schema.TypeString,
							Description: "IoT Topic for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_cutoff": {
							Type:        schema.TypeString,
							Description: "Batch Duration in seconds for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_size": {
							Type:        schema.TypeString,
							Description: "Batch Size for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			triggerTypeMessageQueue: {
				Type:        schema.TypeList,
				Description: "[Message Queue](https://yandex.cloud/docs/functions/concepts/trigger/ymq-trigger) settings definition for Yandex Cloud Functions Trigger, if present.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:        schema.TypeString,
							Description: "Message Queue ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Description: "Message Queue Service Account ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_cutoff": {
							Type:        schema.TypeString,
							Description: "Batch Duration in seconds for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_size": {
							Type:        schema.TypeString,
							Description: "Batch Size for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"visibility_timeout": {
							Type:        schema.TypeString,
							Description: "Visibility timeout for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			triggerTypeObjectStorage: {
				Type:        schema.TypeList,
				Description: "[Object Storage](https://yandex.cloud/docs/functions/concepts/trigger/os-trigger) settings definition for Yandex Cloud Functions Trigger, if present.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_id": {
							Type:        schema.TypeString,
							Description: "Object Storage Bucket ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"prefix": {
							Type:        schema.TypeString,
							Description: "Prefix for Object Storage for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"suffix": {
							Type:        schema.TypeString,
							Description: "Suffix for Object Storage for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"create": {
							Type:        schema.TypeBool,
							Description: "Boolean flag for setting `create` event for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"update": {
							Type:        schema.TypeBool,
							Description: "Boolean flag for setting `update` event for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"delete": {
							Type:        schema.TypeBool,
							Description: "Boolean flag for setting `delete` event for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_cutoff": {
							Type:        schema.TypeString,
							Description: "Batch Duration in seconds for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
						"batch_size": {
							Type:        schema.TypeString,
							Description: "Batch Size for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			triggerTypeContainerRegistry: {
				Type:        schema.TypeList,
				Description: "[Container Registry](https://yandex.cloud/docs/functions/concepts/trigger/cr-trigger) settings definition for Yandex Cloud Functions Trigger, if present.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"registry_id": {
							Type:        schema.TypeString,
							Description: "Container Registry ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"image_name": {
							Type:        schema.TypeString,
							Description: "Image name filter setting for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"tag": {
							Type:        schema.TypeString,
							Description: "Image tag filter setting for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"create_image": {
							Type:        schema.TypeBool,
							Description: "Boolean flag for setting `create image` event for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"delete_image": {
							Type:        schema.TypeBool,
							Description: "Boolean flag for setting `delete image` event for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"create_image_tag": {
							Type:        schema.TypeBool,
							Description: "Boolean flag for setting `create image tag` event for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"delete_image_tag": {
							Type:        schema.TypeBool,
							Description: "Boolean flag for setting `delete image tag` event for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_cutoff": {
							Type:        schema.TypeString,
							Description: "Batch Duration in seconds for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
						"batch_size": {
							Type:        schema.TypeString,
							Description: "Batch Size for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			triggerTypeYDS: {
				Type:        schema.TypeList,
				Description: "[Data Streams](https://yandex.cloud/docs/functions/concepts/trigger/data-streams-trigger) settings definition for Yandex Cloud Functions Trigger, if present.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_name": {
							Type:        schema.TypeString,
							Description: "Stream name for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"database": {
							Type:        schema.TypeString,
							Description: "Stream database for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"suffix": {
							Type:        schema.TypeString,
							Description: "Suffix for Object Storage for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"service_account_id": {
							Type:        schema.TypeBool,
							Description: "Service account ID to access data stream for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_cutoff": {
							Type:        schema.TypeString,
							Description: "Batch Duration in seconds for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
						"batch_size": {
							Type:        schema.TypeString,
							Description: "Batch Size for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			triggerTypeMail: {
				Type:        schema.TypeList,
				Description: "[Mail](https://yandex.cloud/docs/functions/concepts/trigger/mail-trigger) settings definition for Yandex Cloud Functions Trigger, if present.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachments_bucket_id": {
							Type:        schema.TypeString,
							Description: "Object Storage Bucket ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Description: "Service account ID to access object storage for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_cutoff": {
							Type:        schema.TypeString,
							Description: "Batch Duration in seconds for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
						"batch_size": {
							Type:        schema.TypeString,
							Description: "Batch Size for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			triggerTypeTimer: {
				Type:        schema.TypeList,
				Description: "[Timer](https://yandex.cloud/docs/functions/concepts/trigger/timer) settings definition for Yandex Cloud Functions Trigger, if present.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cron_expression": {
							Type:        schema.TypeString,
							Description: "Cron expression for timer for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
						"payload": {
							Type:        schema.TypeString,
							Description: "Payload to be passed to function.",
							Computed:    true,
						},
					},
				},
			},

			triggerTypeLogGroup: {
				Type:        schema.TypeList,
				Description: "Deprecated Logging settings definition for Yandex Cloud Functions Trigger. Please, use logging instead.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_ids": {
							Type:        schema.TypeSet,
							Description: "Log group IDs for Yandex Cloud Functions Trigger.",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},

						"batch_cutoff": {
							Type:        schema.TypeString,
							Description: "Batch Duration in seconds for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_size": {
							Type:        schema.TypeString,
							Description: "Batch Size for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			triggerTypeLogging: {
				Type:        schema.TypeList,
				Description: "[Logging](https://yandex.cloud/docs/functions/concepts/trigger/cloud-logging-trigger) settings definition for Yandex Cloud Functions Trigger, if present.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Type:        schema.TypeString,
							Description: "Logging group ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"resource_ids": {
							Type:        schema.TypeSet,
							Description: "Resource ID filter setting for Yandex Cloud Functions Trigger.",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							MinItems:    0,
						},

						"resource_types": {
							Type:        schema.TypeSet,
							Description: "Resource type filter setting for Yandex Cloud Functions Trigger.",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							MinItems:    0,
						},

						"levels": {
							Type:        schema.TypeSet,
							Description: "Logging level filter setting for Yandex Cloud Functions Trigger.",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							MinItems:    0,
						},

						"stream_names": {
							Type:        schema.TypeSet,
							Description: "Logging stream name filter setting for Yandex Cloud Functions Trigger.",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							MinItems:    0,
						},

						"batch_cutoff": {
							Type:        schema.TypeString,
							Description: "Batch Duration in seconds for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"batch_size": {
							Type:        schema.TypeString,
							Description: "Batch Size for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			"function": {
				Type:        schema.TypeList,
				Description: "[Yandex Cloud Function](https://yandex.cloud/docs/functions/concepts/function) settings definition for Yandex Cloud Functions Trigger.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Yandex Cloud Function ID.",
							Computed:    true,
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Description: "Service account ID for Yandex Cloud Function.",
							Computed:    true,
						},

						"tag": {
							Type:        schema.TypeString,
							Description: "Tag for Yandex Cloud Function for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"retry_attempts": {
							Type:        schema.TypeString,
							Description: "Retry attempts for Yandex Cloud Function for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"retry_interval": {
							Type:        schema.TypeString,
							Description: "Retry interval in seconds for Yandex Cloud Function for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			"container": {
				Type:        schema.TypeList,
				Description: "[Yandex Cloud Serverless Container](https://yandex.cloud/docs/serverless-containers/concepts/container) settings definition for Yandex Cloud Functions Trigger.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Yandex Cloud Serverless Container ID for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Description: "Service account ID for Yandex Cloud Serverless Container for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"path": {
							Type:        schema.TypeString,
							Description: "Path for Yandex Cloud Serverless Container for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"retry_attempts": {
							Type:        schema.TypeString,
							Description: "Retry attempts for Yandex Cloud Serverless Container for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},

						"retry_interval": {
							Type:        schema.TypeString,
							Description: "Retry interval in seconds for Yandex Cloud Serverless Container for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			"dlq": {
				Type:        schema.TypeList,
				Description: "Dead Letter Queue (DLQ) settings definition for Yandex Cloud Functions Trigger.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:        schema.TypeString,
							Description: "ID of Dead Letter Queue for Trigger (Queue ARN).",
							Computed:    true,
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Description: "Service Account ID for Dead Letter Queue for Yandex Cloud Functions Trigger.",
							Computed:    true,
						},
					},
				},
			},

			"workflow": {
				Type:        schema.TypeList,
				Description: "Workflows settings definition for Yandex Cloud Functions Trigger.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Workflow ID.",
							Computed:    true,
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Description: "Service account ID for Workflows.",
							Computed:    true,
						},

						"retry_attempts": {
							Type:        schema.TypeString,
							Description: "Retry attempts for Workflows.",
							Computed:    true,
						},

						"retry_interval": {
							Type:        schema.TypeString,
							Description: "Retry interval in seconds for Workflows.",
							Computed:    true,
						},
					},
				},
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexFunctionTriggerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "trigger_id", "name")
	if err != nil {
		return err
	}

	triggerID := d.Get("trigger_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		triggerID, err = resolveObjectID(ctx, config, d, sdkresolvers.TriggerResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Yandex Cloud Functions Trigger by name: %v", err)
		}
	}

	req := triggers.GetTriggerRequest{
		TriggerId: triggerID,
	}

	trig, err := config.sdk.Serverless().Triggers().Trigger().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Functions Trigger %q", d.Id()))
	}

	d.SetId(trig.Id)
	d.Set("trigger_id", trig.Id)
	return flattenYandexFunctionTrigger(d, trig)
}
