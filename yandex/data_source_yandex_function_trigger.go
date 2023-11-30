package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexFunctionTrigger() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexFunctionTriggerRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"trigger_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			triggerTypeIoT: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"registry_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"device_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"topic": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"batch_size": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			triggerTypeMessageQueue: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"batch_size": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"visibility_timeout": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			triggerTypeObjectStorage: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"suffix": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"create": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"update": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"delete": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			triggerTypeContainerRegistry: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"registry_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"image_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"create_image": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"delete_image": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"create_image_tag": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"delete_image_tag": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			triggerTypeYDS: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"database": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"suffix": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"service_account_id": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			triggerTypeMail: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachments_bucket_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			triggerTypeTimer: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cron_expression": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"payload": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			triggerTypeLogGroup: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"batch_size": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			triggerTypeLogging: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"resource_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 0,
						},

						"resource_types": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 0,
						},

						"levels": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 0,
						},

						"stream_names": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 0,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"batch_size": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"function": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"retry_attempts": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"retry_interval": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"container": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"retry_attempts": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"retry_interval": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"dlq": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
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
