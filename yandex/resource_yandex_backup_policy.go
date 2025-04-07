package yandex

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexBackupDefaultTimeout                            = 5 * time.Minute
	yandexBackupPolicySchedulingIntervalMaxAvailableValue = 1 * 60 * 60 * 24 * 9999 // max is 9999 days
)

func resourceYandexBackupPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of [Yandex Cloud Backup Policy](https://yandex.cloud/docs/backup/concepts/policy).\n\n~> Cloud Backup Provider must be activated in order to manipulate with policies. Active it either by UI Console or by `yc` command.\n\n## Defined types\n\n### interval_type\n\n A string type, that accepts values in the format of: `number` + `time type`, where `time type` might be:\n* `s` — seconds\n* `m` — minutes\n* `h` — hours\n* `d` — days\n* `w` — weekdays\n* `M` — months\n\nExample of interval value: `5m`, `10d`, `2M`, `5w`\n\n### day_type\n\nA string type, that accepts the following values: `ALWAYS_INCREMENTAL`, `ALWAYS_FULL`, `WEEKLY_FULL_DAILY_INCREMENTAL`, `WEEKLY_INCREMENTAL`.\n\n### backup_set_type\n\n`TYPE_AUTO`, `TYPE_FULL`, `TYPE_INCREMENTAL`, `TYPE_DIFFERENTIAL`.",

		CreateContext: resourceYandexBackupPolicyCreate,
		ReadContext:   resourceYandexBackupPolicyRead,
		UpdateContext: resourceYandexBackupPolicyUpdate,
		DeleteContext: resourceYandexBackupPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexBackupDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexBackupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexBackupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexBackupDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},

			"compression": {
				Type:         schema.TypeString,
				Description:  "Archive compression level. Affects CPU. Available values: `NORMAL`, `HIGH`, `MAX`, `OFF`. Default: `NORMAL`.",
				Optional:     true,
				Default:      resourceYandexBackupCompressionValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupCompressionValues, false),
			},

			"format": {
				Type:         schema.TypeString,
				Description:  "Format of the backup. It's strongly recommend to leave this option empty or `AUTO`. Available values: `AUTO`, `VERSION_11`, `VERSION_12`.",
				Optional:     true,
				Default:      resourceYandexBackupPolicyFormatValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupPolicyFormatValues, false),
			},

			"multi_volume_snapshotting_enabled": {
				Type:        schema.TypeBool,
				Description: "If true, snapshots of multiple volumes will be taken simultaneously. Default `true`.",
				Optional:    true,
				Default:     true,
			},

			"preserve_file_security_settings": {
				Type:        schema.TypeBool,
				Description: "Preserves file security settings. It's better to set this option to true. Default `true`.",
				Optional:    true,
				Default:     true,
			},

			"reattempts": {
				Type:        schema.TypeSet,
				Description: "Amount of reattempts that should be performed while trying to make backup at the host.",
				MaxItems:    1,
				Required:    true,
				Elem:        resourceYandexBackupPolicyRetryConfigurationResource(),
				Set:         storageBucketS3SetFunc("enabled", "interval", "max_attempts"),
			},

			"vm_snapshot_reattempts": {
				Type:        schema.TypeSet,
				Description: "Amount of reattempts that should be performed while trying to make snapshot.",
				MaxItems:    1,
				Required:    true,
				Elem:        resourceYandexBackupPolicyRetryConfigurationResource(),
			},

			"silent_mode_enabled": {
				Type:        schema.TypeBool,
				Description: "If true, a user interaction will be avoided when possible. Default `true`.",
				Optional:    true,
				Default:     true,
			},

			"splitting_bytes": {
				Type:        schema.TypeString,
				Description: "Determines the size to split backups. It's better to leave this option unchanged. Default `9223372036854775807`.",
				Optional:    true,
				Default:     "9223372036854775807", // almost max int
			},

			"vss_provider": {
				Type:         schema.TypeString,
				Description:  "Settings for the volume shadow copy service. Available values are: `NATIVE`, `TARGET_SYSTEM_DEFINED`. Default `NATIVE`.",
				Optional:     true,
				Default:      resourceYandexBackupVSSProviderValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupVSSProviderValues, false),
			},

			"archive_name": {
				Type:        schema.TypeString,
				Description: "The name of generated archives. Default `[Machine Name]-[Plan ID]-[Unique ID]a`.",
				Optional:    true,
				Default:     "[Machine Name]-[Plan ID]-[Unique ID]a",
				ValidateFunc: func(v any, _ string) (warnings []string, errs []error) {
					value := v.(string)
					if resourceYandexBackupBadArchiveNameTemplate.MatchString(value) {
						errs = append(errs, fmt.Errorf("archive_name must not end with variable or digit"))
					}

					return warnings, errs
				},
			},

			"performance_window_enabled": {
				Type:        schema.TypeBool,
				Description: "Time windows for performance limitations of backup. Default `false`.",
				Optional:    true,
				Default:     false,
			},

			"retention": {
				Type:        schema.TypeSet,
				Description: "Retention policy for backups. Allows to setup backups lifecycle.",
				Required:    true,
				MaxItems:    1,
				Elem:        resourceYandexBacupPolicyRetentionSchema(),
				Set:         storageBucketS3SetFunc("after_backup", "rules"),
			},

			"scheduling": {
				Type:        schema.TypeSet,
				Description: "Schedule settings for creating backups on the host.",
				Required:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem:        resourceYandexBackupPolicySchedulingResource(),
			},

			"cbt": {
				Type:         schema.TypeString,
				Description:  "Configuration of Changed Block Tracking. Available values are: `USE_IF_ENABLED`, `ENABLED_AND_USE`, `DO_NOT_USE`. Default `DO_NOT_USE`.",
				Optional:     true,
				Default:      resourceYandexBackupCBTValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupCBTValues, false),
			},

			"file_filters": {
				Type:        schema.TypeList,
				Description: "File filters to specify masks of files to backup or to exclude of backuping.",
				Optional:    true,
				MaxItems:    1,
				Elem:        resourceYandexBacupPolicyFileFiltersSchema(),
			},

			"fast_backup_enabled": {
				Type:        schema.TypeBool,
				Description: "If true, determines whether a file has changed by the file size and timestamp. Otherwise, the entire file contents are compared to those stored in the backup.",
				Optional:    true,
				Default:     true,
			},

			"quiesce_snapshotting_enabled": {
				Type:        schema.TypeBool,
				Description: "If true, a quiesced snapshot of the virtual machine will be taken. Default `false`.",
				Optional:    true,
				Default:     false,
			},

			// COMPUTED ONLY VALUES

			"enabled": {
				Type:        schema.TypeBool,
				Description: "If this field is true, it means that the policy is enabled.",
				Optional:    false,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Optional:    false,
				Computed:    true,
			},

			"updated_at": {
				Type:        schema.TypeString,
				Description: "The update timestamp of the resource.",
				Optional:    false,
				Computed:    true,
			},
		},
	}
}

func resourceYandexBackupPolicyRetryConfigurationResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Enable flag. Default `true`.",
				Optional:    true,
				Default:     true,
			},

			"interval": {
				Type:         schema.TypeString,
				Description:  "Retry interval. See `interval_type` for available values. Default: `5m`.",
				Optional:     true,
				Default:      "5m",
				ValidateFunc: resourceYandexBackupIntervalValidationFunc,
			},

			"max_attempts": {
				Type:        schema.TypeInt,
				Description: "Maximum number of attempts before throwing an error. Default `5`.",
				Optional:    true,
				Default:     5,
			},
		},
	}
}

func resourceYandexBackupRetentionRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_age": {
				Type:         schema.TypeString,
				Description:  "Deletes backups that older than `max_age`. Exactly one of `max_count` or `max_age` should be set.",
				Optional:     true,
				ValidateFunc: resourceYandexBackupIntervalValidationFunc,
			},

			"max_count": {
				Type:        schema.TypeInt,
				Description: "Deletes backups if it's count exceeds `max_count`. Exactly one of `max_count` or `max_age` should be set.",
				Optional:    true,
			},

			"repeat_period": {
				Type:        schema.TypeList,
				Description: "Possible types: `REPEATE_PERIOD_UNSPECIFIED`, `HOURLY`, `DAILY`, `WEEKLY`, `MONTHLY`. Specifies repeat period of the backupset.",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(resourceYandexBackupRepeatPeriodValues, false),
				},
			},
		},
	}
}

func resourceYandexBacupPolicyFileFiltersSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"exclusion_masks": {
				Type: schema.TypeList,

				Description: "Do not backup files that match the following criteria.",

				Optional: true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"inclusion_masks": {
				Type: schema.TypeList,

				Description: "Backup only files that match the following criteria.",

				Optional: true,

				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceYandexBacupPolicyRetentionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"after_backup": {
				Type:        schema.TypeBool,
				Description: "Defines whether retention rule applies after creating backup or before.",
				Optional:    true,
				Default:     true,
			},

			"rules": {
				Type:        schema.TypeSet,
				Description: "A list of retention rules.",
				Optional:    true,
				Elem:        resourceYandexBackupRetentionRuleSchema(),
				Set:         storageBucketS3SetFunc("max_age", "max_count", "repeat_period"),
			},
		},
	}
}

func resourceYandexBackupPolicySchedulingResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"execute_by_interval": {
				Type:         schema.TypeInt,
				Description:  " Perform backup by interval, since last backup of the host. Maximum value is: 9999 days. See `interval_type` for available values. Exactly on of options should be set: `execute_by_interval` or `execute_by_time`.",
				Optional:     true,
				ValidateFunc: resourceYandexBackupSchedulingIntervalValidateFunc,
				Deprecated:   fieldDeprecatedForAnother("execute_by_interval", "backup_sets"),
			},

			"execute_by_time": {
				Type:        schema.TypeSet,
				Description: "Perform backup periodically at specific time. Exactly on of options should be set: `execute_by_interval` or `execute_by_time`.",
				Optional:    true,
				Set:         storageBucketS3SetFunc("weekdays", "repeat_at", "repeat_every", "monthdays", "include_last_day_of_month", "months", "type"),
				Elem:        resourceYandexBackupPolicySchedulingRuleTimeResource(),
				Deprecated:  fieldDeprecatedForAnother("execute_by_time", "backup_sets"),
			},

			"backup_sets": {
				Type:        schema.TypeSet,
				Description: "A list of schedules with backup sets that compose the whole scheme.",
				Optional:    true,
				MinItems:    1,
				Set:         storageBucketS3SetFunc("execute_by_interval", "execute_by_time", "type"),
				Elem:        resourceYandexBackupPolicySchedulingBackupSetResource(),
			},

			"enabled": {
				Type:        schema.TypeBool,
				Description: "Enables or disables scheduling. Default `true`.",
				Optional:    true,
				Default:     true,
			},

			"max_parallel_backups": {
				Type:        schema.TypeInt,
				Description: "Maximum number of backup processes allowed to run in parallel. 0 for unlimited. Default `0`.",
				Optional:    true,
				Default:     0,
			},

			"random_max_delay": {
				Type:         schema.TypeString,
				Description:  "Configuration of the random delay between the execution of parallel tasks. See `interval_type` for available values. Default `30m`.",
				Optional:     true,
				Default:      "30m",
				ValidateFunc: resourceYandexBackupIntervalValidationFunc,
			},

			"scheme": {
				Type:         schema.TypeString,
				Description:  "Scheme of the backups. Available values are: `ALWAYS_INCREMENTAL`, `ALWAYS_FULL`, `WEEKLY_FULL_DAILY_INCREMENTAL`, `WEEKLY_INCREMENTAL`. Default `ALWAYS_INCREMENTAL`.",
				Optional:     true,
				Default:      resourceYandexBackupTypeValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupTypeValues, false),
			},

			"weekly_backup_day": {
				Type:         schema.TypeString,
				Description:  "A day of week to start weekly backups. See `day_type` for available values. Default `MONDAY`.",
				Optional:     true,
				Default:      resourceYandexBackupDayValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupDayValues, false),
			},
		},
	}
}

func resourceYandexBackupPolicySchedulingBackupSetResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"execute_by_interval": {
				Type:         schema.TypeInt,
				Description:  "Perform backup by interval, since last backup of the host. Maximum value is: 9999 days. See `interval_type` for available values. Exactly on of options should be set: `execute_by_interval` or `execute_by_time`.",
				Optional:     true,
				ValidateFunc: resourceYandexBackupSchedulingIntervalValidateFunc,
			},

			"execute_by_time": {
				Type:        schema.TypeSet,
				Description: "Perform backup periodically at specific time. Exactly on of options should be set: `execute_by_interval` or `execute_by_time`.",
				Optional:    true,
				Set:         storageBucketS3SetFunc("weekdays", "repeat_at", "repeat_every", "monthdays", "include_last_day_of_month", "months", "type"),
				Elem:        resourceYandexBackupPolicySchedulingRuleTimeResource(),
			},

			"type": {
				Type:         schema.TypeString,
				Description:  "BackupSet type. See `backup_set_type` for available values. Default `TYPE_AUTO`.",
				Optional:     true,
				Default:      resourceYandexBackupSchedulingBackupSetTypeValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupSchedulingBackupSetTypeValues, false),
			},
		},
	}
}

func resourceYandexBackupPolicySchedulingRuleTimeResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"weekdays": {
				Type:        schema.TypeList,
				Description: "List of weekdays when the backup will be applied. Used in `WEEKLY` type.",
				Optional:    true,
				MaxItems:    7,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(resourceYandexBackupDayValues, false),
				},
			},

			"repeat_at": {
				Type:        schema.TypeList,
				Description: "List of time in format `HH:MM` (24-hours format), when the schedule applies.",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: resourceYandexBackupPolicySettingsTimeOfDayValidateFunc,
				},
			},

			"repeat_every": {
				Type:         schema.TypeString,
				Description:  "Frequency of backup repetition. See `interval_type` for available values.",
				Optional:     true,
				ValidateFunc: resourceYandexBackupIntervalValidationFunc,
			},

			"monthdays": {
				Type:        schema.TypeList,
				Description: "List of days when schedule applies. Used in `MONTHLY` type.",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 31),
				},
			},

			"include_last_day_of_month": {
				Type:        schema.TypeBool,
				Description: "If true, schedule will be applied on the last day of month. See `day_type` for available values. Default `true`.",
				Optional:    true,
				Default:     false,
			},

			"months": {
				Type:        schema.TypeList,
				Description: "Set of values. Allowed values form 1 to 12.",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 12),
				},
			},

			"type": {
				Type:         schema.TypeString,
				Description:  "Type of the scheduling. Available values are: `HOURLY`, `DAILY`, `WEEKLY`, `MONTHLY`.",
				ValidateFunc: validation.StringInSlice(resourceYandexBackupRepeatPeriodValues, false),
				Required:     true,
			},
		},
	}
}

func expandBackupPolicySettingsInterval(v any) *backuppb.PolicySettings_Interval {
	value, ok := v.(string)
	if !ok {
		return nil
	}

	if value == "" {
		return nil
	}

	result := resourceYandexBackupIntervalTemplate.FindStringSubmatch(value)
	if len(result) == 0 {
		log.Printf("[WARN] no result parsing interval")
		return nil
	}

	var err error
	out := &backuppb.PolicySettings_Interval{}
	out.Count, err = strconv.ParseInt(result[1], 10, 64)
	if err != nil {
		panic("unexpected digits for interval " + err.Error())
	}

	switch result[2] {
	case "s":
		out.Type = backuppb.PolicySettings_Interval_SECONDS
	case "m":
		out.Type = backuppb.PolicySettings_Interval_MINUTES
	case "h":
		out.Type = backuppb.PolicySettings_Interval_HOURS
	case "d":
		out.Type = backuppb.PolicySettings_Interval_DAYS
	case "w":
		out.Type = backuppb.PolicySettings_Interval_WEEKS
	case "M":
		out.Type = backuppb.PolicySettings_Interval_MONTHS
	default:
		panic("unknown value " + result[1])
	}

	return out
}

func resourceYandexBackupPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.Errorf("getting folder while creating policy: %s", err)
	}

	name := d.Get("name").(string)
	settings, err := expandYandexBackupPolicySettingsFromResource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	request := &backuppb.CreatePolicyRequest{
		FolderId: folderID,
		Name:     name,
		Settings: settings,
	}

	log.Printf("[INFO] starting to create Cloud Backup policy with request %s", request.String())

	operation, err := config.sdk.WrapOperation(config.sdk.Backup().Policy().Create(ctx, request))
	if err != nil {
		return diag.Errorf("requesting API to create Cloud Backup Policy: %s", err)
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return diag.FromErr(err)
	}

	pm, ok := protoMetadata.(*backuppb.CreatePolicyMetadata)
	if !ok {
		return diag.Errorf("unexpected policy metadata type %T", protoMetadata)
	}

	d.SetId(pm.PolicyId)
	log.Printf("[INFO] Created Cloud Backup policy with id=%q", pm.PolicyId)

	if err = operation.Wait(ctx); err != nil {
		return diag.Errorf("waiting for operation completes: %s", err)
	}

	return resourceYandexBackupPolicyRead(ctx, d, meta)
}

func resourceYandexBackupPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*Config)
	id := d.Id()

	log.Printf("[DEBUG] Starting to fetch Cloud Backup policy with id=%q", id)

	policy, err := config.sdk.Backup().Policy().Get(ctx, &backuppb.GetPolicyRequest{
		PolicyId: id,
	})
	if err != nil {
		err = handleNotFoundError(err, d, id)
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Fetched Cloud Backup Policy %s", policy.String())

	if err = flattenBackupPolicy(d, policy); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexBackupPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*Config)
	id := d.Id()

	log.Printf("[DEBUG] Starting to update Cloud Backup policy with id=%q", id)

	request, err := prepareBackupPolicyUpdateRequest(d, config)
	if err != nil {
		err = handleNotFoundError(err, d, id)
		return diag.Errorf("preparing backup policy update request: %s", err)
	}

	log.Printf("[INFO] Starting to update Cloud Backup policy with id=%q and request=%s", id, request.String())

	operation, err := config.sdk.WrapOperation(config.sdk.Backup().Policy().Update(ctx, request))
	if err != nil {
		return diag.Errorf("updating policy: %s", err)
	}

	err = operation.Wait(ctx)
	if err != nil {
		return diag.Errorf("waiting for operation completes: %s", err)
	}

	return resourceYandexBackupPolicyRead(ctx, d, meta)
}

func resourceYandexBackupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*Config)
	id := d.Id()

	log.Printf("[INFO] Starting to delete Cloud Backup policy with id=%q", id)

	operation, err := config.sdk.WrapOperation(config.sdk.Backup().Policy().Delete(ctx, &backuppb.DeletePolicyRequest{
		PolicyId: d.Id(),
	}))
	if err != nil {
		err = handleNotFoundError(err, d, id)
		return diag.FromErr(err)
	}

	err = operation.Wait(ctx)
	if err != nil {
		return diag.Errorf("waiting operation for completes: %s", err)
	}

	return resourceYandexBackupPolicyRead(ctx, d, meta)
}

func expandInt64(v any) []int64 {
	values := v.([]any)
	out := make([]int64, 0, len(values))

	for _, value := range values {
		out = append(out, int64(value.(int)))
	}

	return out
}
