package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
)

func dataSourceYandexBackupPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexBackupPolicyRead,
		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"compression": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"format": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"multi_volume_snapshotting_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"preserve_file_security_settings": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"reattempts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"interval": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max_attempts": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Set: storageBucketS3SetFunc("enabled", "interval", "max_attempts"),
			},

			"vm_snapshot_reattempts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"interval": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max_attempts": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"silent_mode_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"splitting_bytes": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"vss_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"archive_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"performance_window_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"retention": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after_backup": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"rules": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     resourceYandexBackupRetentionRuleSchema(),
							Set:      storageBucketS3SetFunc("max_age", "max_count", "repeat_period"),
						},
					},
				},
				Set: storageBucketS3SetFunc("after_backup", "rules"),
			},

			"scheduling": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_sets": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Set:      storageBucketS3SetFunc("execute_by_interval", "execute_by_time", "type"),
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"execute_by_interval": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"execute_by_time": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"weekdays": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},

												"repeat_at": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},

												"repeat_every": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"monthdays": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeInt,
													},
												},

												"include_last_day_of_month": {
													Type:     schema.TypeBool,
													Computed: true,
												},

												"months": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeInt,
													},
												},

												"type": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
										Set: storageBucketS3SetFunc("weekdays", "repeat_at", "repeat_every", "monthdays", "include_last_day_of_month", "months", "type"),
									},

									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"max_parallel_backups": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"random_max_delay": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"scheme": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"weekly_backup_day": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"cbt": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"fast_backup_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"quiesce_snapshotting_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
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
