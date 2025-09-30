package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexBackupPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Backup Policy. For more information, see [the official documentation](https://yandex.cloud/docs/backup/concepts/policy).\n\n~> One of `policy_id` or `name` should be specified.\n\n~> In case you use `name`, an error will occur if two policies with the same name exist. In this case, rename the policy or use the `policy_id`.\n",
		ReadContext: dataSourceYandexBackupPolicyRead,
		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:        schema.TypeString,
				Description: "ID of the policy.",
				Optional:    true,
				Computed:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
			},

			"compression": {
				Type:        schema.TypeString,
				Description: resourceYandexBackupPolicy().Schema["compression"].Description,
				Computed:    true,
			},

			"format": {
				Type:        schema.TypeString,
				Description: resourceYandexBackupPolicy().Schema["format"].Description,
				Computed:    true,
			},

			"multi_volume_snapshotting_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexBackupPolicy().Schema["multi_volume_snapshotting_enabled"].Description,
				Computed:    true,
			},

			"reattempts": {
				Type:        schema.TypeSet,
				Description: resourceYandexBackupPolicy().Schema["reattempts"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: resourceYandexBackupPolicyRetryConfigurationResource().Schema["enabled"].Description,
							Computed:    true,
						},
						"interval": {
							Type:        schema.TypeString,
							Description: resourceYandexBackupPolicyRetryConfigurationResource().Schema["interval"].Description,
							Computed:    true,
						},
						"max_attempts": {
							Type:        schema.TypeInt,
							Description: resourceYandexBackupPolicyRetryConfigurationResource().Schema["max_attempts"].Description,
							Computed:    true,
						},
					},
				},
				Set: storageBucketS3SetFunc("enabled", "interval", "max_attempts"),
			},

			"vm_snapshot_reattempts": {
				Type:        schema.TypeSet,
				Description: resourceYandexBackupPolicy().Schema["vm_snapshot_reattempts"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: resourceYandexBackupPolicyRetryConfigurationResource().Schema["enabled"].Description,
							Computed:    true,
						},
						"interval": {
							Type:        schema.TypeString,
							Description: resourceYandexBackupPolicyRetryConfigurationResource().Schema["interval"].Description,
							Computed:    true,
						},
						"max_attempts": {
							Type:        schema.TypeInt,
							Description: resourceYandexBackupPolicyRetryConfigurationResource().Schema["max_attempts"].Description,
							Computed:    true,
						},
					},
				},
			},

			"silent_mode_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexBackupPolicy().Schema["silent_mode_enabled"].Description,
				Computed:    true,
			},

			"splitting_bytes": {
				Type:        schema.TypeString,
				Description: resourceYandexBackupPolicy().Schema["splitting_bytes"].Description,
				Computed:    true,
			},

			"vss_provider": {
				Type:        schema.TypeString,
				Description: resourceYandexBackupPolicy().Schema["vss_provider"].Description,
				Computed:    true,
			},

			"archive_name": {
				Type:        schema.TypeString,
				Description: resourceYandexBackupPolicy().Schema["archive_name"].Description,
				Computed:    true,
			},

			"performance_window_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexBackupPolicy().Schema["performance_window_enabled"].Description,
				Computed:    true,
			},

			"retention": {
				Type:        schema.TypeSet,
				Description: resourceYandexBackupPolicy().Schema["retention"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after_backup": {
							Type:        schema.TypeBool,
							Description: resourceYandexBacupPolicyRetentionSchema().Schema["after_backup"].Description,
							Computed:    true,
						},
						"rules": {
							Type:        schema.TypeSet,
							Description: resourceYandexBacupPolicyRetentionSchema().Schema["rules"].Description,
							Computed:    true,
							Elem:        resourceYandexBackupRetentionRuleSchema(),
							Set:         storageBucketS3SetFunc("max_age", "max_count", "repeat_period"),
						},
					},
				},
				Set: storageBucketS3SetFunc("after_backup", "rules"),
			},

			"scheduling": {
				Type:        schema.TypeSet,
				Description: resourceYandexBackupPolicy().Schema["scheduling"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_sets": {
							Type:        schema.TypeSet,
							Description: resourceYandexBackupPolicySchedulingResource().Schema["backup_sets"].Description,
							Optional:    true,
							MinItems:    1,
							Set:         storageBucketS3SetFunc("execute_by_interval", "execute_by_time", "type"),
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"execute_by_interval": {
										Type:        schema.TypeInt,
										Description: resourceYandexBackupPolicySchedulingBackupSetResource().Schema["execute_by_interval"].Description,
										Computed:    true,
									},
									"execute_by_time": {
										Type:        schema.TypeSet,
										Description: resourceYandexBackupPolicySchedulingBackupSetResource().Schema["execute_by_time"].Description,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"weekdays": {
													Type:        schema.TypeList,
													Description: resourceYandexBackupPolicySchedulingRuleTimeResource().Schema["weekdays"].Description,
													Computed:    true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},

												"repeat_at": {
													Type:        schema.TypeList,
													Description: resourceYandexBackupPolicySchedulingRuleTimeResource().Schema["repeat_at"].Description,
													Computed:    true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},

												"repeat_every": {
													Type:        schema.TypeString,
													Description: resourceYandexBackupPolicySchedulingRuleTimeResource().Schema["repeat_every"].Description,
													Computed:    true,
												},

												"monthdays": {
													Type:        schema.TypeList,
													Description: resourceYandexBackupPolicySchedulingRuleTimeResource().Schema["monthdays"].Description,
													Computed:    true,
													Elem: &schema.Schema{
														Type: schema.TypeInt,
													},
												},

												"include_last_day_of_month": {
													Type:        schema.TypeBool,
													Description: resourceYandexBackupPolicySchedulingRuleTimeResource().Schema["include_last_day_of_month"].Description,
													Computed:    true,
												},

												"months": {
													Type:        schema.TypeList,
													Description: resourceYandexBackupPolicySchedulingRuleTimeResource().Schema["months"].Description,
													Computed:    true,
													Elem: &schema.Schema{
														Type: schema.TypeInt,
													},
												},

												"run_later": {
													Type:        schema.TypeBool,
													Description: resourceYandexBackupPolicySchedulingRuleTimeResource().Schema["run_later"].Description,
													Computed:    true,
												},

												"type": {
													Type:        schema.TypeString,
													Description: resourceYandexBackupPolicySchedulingRuleTimeResource().Schema["type"].Description,
													Computed:    true,
												},
											},
										},
										Set: storageBucketS3SetFunc("weekdays", "repeat_at", "repeat_every", "monthdays", "include_last_day_of_month", "months", "type", "run_later"),
									},

									"type": {
										Type:        schema.TypeString,
										Description: resourceYandexBackupPolicySchedulingBackupSetResource().Schema["type"].Description,
										Computed:    true,
									},
								},
							},
						},

						"enabled": {
							Type:        schema.TypeBool,
							Description: resourceYandexBackupPolicySchedulingResource().Schema["enabled"].Description,
							Computed:    true,
						},

						"max_parallel_backups": {
							Type:        schema.TypeInt,
							Description: resourceYandexBackupPolicySchedulingResource().Schema["max_parallel_backups"].Description,
							Computed:    true,
						},

						"random_max_delay": {
							Type:        schema.TypeString,
							Description: resourceYandexBackupPolicySchedulingResource().Schema["random_max_delay"].Description,
							Computed:    true,
						},

						"scheme": {
							Type:        schema.TypeString,
							Description: resourceYandexBackupPolicySchedulingResource().Schema["scheme"].Description,
							Computed:    true,
						},

						"weekly_backup_day": {
							Type:        schema.TypeString,
							Description: resourceYandexBackupPolicySchedulingResource().Schema["weekly_backup_day"].Description,
							Computed:    true,
						},
					},
				},
			},

			"cbt": {
				Type:        schema.TypeString,
				Description: resourceYandexBackupPolicy().Schema["cbt"].Description,
				Computed:    true,
			},

			"file_filters": {
				Type:        schema.TypeList,
				Description: resourceYandexBackupPolicy().Schema["file_filters"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclusion_masks": {
							Type:        schema.TypeList,
							Description: resourceYandexBacupPolicyFileFiltersSchema().Schema["exclusion_masks"].Description,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"inclusion_masks": {
							Type:        schema.TypeList,
							Description: resourceYandexBacupPolicyFileFiltersSchema().Schema["inclusion_masks"].Description,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},

			"fast_backup_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexBackupPolicy().Schema["fast_backup_enabled"].Description,
				Computed:    true,
			},

			"sector_by_sector": {
				Type:        schema.TypeBool,
				Description: resourceYandexBackupPolicy().Schema["sector_by_sector"].Description,
				Computed:    true,
			},

			"validation_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexBackupPolicy().Schema["validation_enabled"].Description,
				Computed:    true,
			},

			"lvm_snapshotting_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexBackupPolicy().Schema["lvm_snapshotting_enabled"].Description,
				Computed:    true,
			},

			"enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexBackupPolicy().Schema["enabled"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"updated_at": {
				Type:        schema.TypeString,
				Description: resourceYandexBackupPolicy().Schema["updated_at"].Description,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexBackupPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "policy_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	var policy *backuppb.Policy
	var resourceName string
	policyID := d.Get("policy_id").(string)
	if policyName, policyNameOk := d.GetOk("name"); policyNameOk {
		resourceName = policyName.(string)
		policy, err = getPolicyByName(ctx, config, resourceName)
	} else {
		resourceName = d.Get("policy_id").(string)
		policy, err = config.sdk.Backup().Policy().Get(ctx, &backuppb.GetPolicyRequest{
			PolicyId: policyID,
		})
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, resourceName))
	}

	d.Set("name", policy.Name)
	d.Set("policy_id", policy.Id)

	if err = flattenBackupPolicy(d, policy); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
