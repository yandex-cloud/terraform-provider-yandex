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
)

const (
	yandexBackupDefaultTimeout                            = 5 * time.Minute
	yandexBackupPolicySchedulingIntervalMaxAvailableValue = 1 * 60 * 60 * 24 * 9999 // max is 9999 days
)

func resourceYandexBackupPolicy() *schema.Resource {
	return &schema.Resource{
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
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"compression": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      resourceYandexBackupCompressionValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupCompressionValues, false),
			},

			"format": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      resourceYandexBackupPolicyFormatValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupPolicyFormatValues, false),
			},

			"multi_volume_snapshotting_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"preserve_file_security_settings": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"reattempts": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Required: true,
				Elem:     resourceYandexBackupPolicyRetryConfigurationResource(),
				Set:      storageBucketS3SetFunc("enabled", "interval", "max_attempts"),
			},

			"vm_snapshot_reattempts": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Required: true,
				Elem:     resourceYandexBackupPolicyRetryConfigurationResource(),
			},

			"silent_mode_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"splitting_bytes": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "9223372036854775807", // almost max int
			},

			"vss_provider": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      resourceYandexBackupVSSProviderValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupVSSProviderValues, false),
			},

			"archive_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "[Machine Name]-[Plan ID]-[Unique ID]a",
				ValidateFunc: func(v any, _ string) (warnings []string, errs []error) {
					value := v.(string)
					if resourceYandexBackupBadArchiveNameTemplate.MatchString(value) {
						errs = append(errs, fmt.Errorf("archive_name must not end with variable or digit"))
					}

					return warnings, errs
				},
			},

			"performance_window_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"retention": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem:     resourceYandexBacupPolicyRetentionSchema(),
				Set:      storageBucketS3SetFunc("after_backup", "rules"),
			},

			"scheduling": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     resourceYandexBackupPolicySchedulingResource(),
			},

			"cbt": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      resourceYandexBackupCBTValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupCBTValues, false),
			},

			"fast_backup_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"quiesce_snapshotting_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// COMPUTED ONLY VALUES

			"enabled": {
				Type:     schema.TypeBool,
				Optional: false,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},

			"updated_at": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
		},
	}
}

func resourceYandexBackupPolicyRetryConfigurationResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"interval": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "5m",
				ValidateFunc: resourceYandexBackupIntervalValidationFunc,
			},

			"max_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5,
			},
		},
	}
}

func resourceYandexBackupRetentionRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_age": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: resourceYandexBackupIntervalValidationFunc,
			},

			"max_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"repeat_period": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(resourceYandexBackupRepeatPeriodValues, false),
				},
			},
		},
	}
}

func resourceYandexBacupPolicyRetentionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"after_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"rules": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     resourceYandexBackupRetentionRuleSchema(),
				Set:      storageBucketS3SetFunc("max_age", "max_count", "repeat_period"),
			},
		},
	}
}

func resourceYandexBackupPolicySchedulingResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"execute_by_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: resourceYandexBackupSchedulingIntervalValidateFunc,
				Deprecated:   fieldDeprecatedForAnother("execute_by_interval", "backup_sets"),
			},

			"execute_by_time": {
				Type:       schema.TypeSet,
				Optional:   true,
				Set:        storageBucketS3SetFunc("weekdays", "repeat_at", "repeat_every", "monthdays", "include_last_day_of_month", "months", "type"),
				Elem:       resourceYandexBackupPolicySchedulingRuleTimeResource(),
				Deprecated: fieldDeprecatedForAnother("execute_by_time", "backup_sets"),
			},

			"backup_sets": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Set:      storageBucketS3SetFunc("execute_by_interval", "execute_by_time", "type"),
				Elem:     resourceYandexBackupPolicySchedulingBackupSetResource(),
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"max_parallel_backups": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"random_max_delay": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "30m",
				ValidateFunc: resourceYandexBackupIntervalValidationFunc,
			},

			"scheme": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      resourceYandexBackupTypeValues[0],
				ValidateFunc: validation.StringInSlice(resourceYandexBackupTypeValues, false),
			},

			"weekly_backup_day": {
				Type:         schema.TypeString,
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
				Optional:     true,
				ValidateFunc: resourceYandexBackupSchedulingIntervalValidateFunc,
			},

			"execute_by_time": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      storageBucketS3SetFunc("weekdays", "repeat_at", "repeat_every", "monthdays", "include_last_day_of_month", "months", "type"),
				Elem:     resourceYandexBackupPolicySchedulingRuleTimeResource(),
			},

			"type": {
				Type:         schema.TypeString,
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
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 7,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(resourceYandexBackupDayValues, false),
				},
			},

			"repeat_at": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: resourceYandexBackupPolicySettingsTimeOfDayValidateFunc,
				},
			},

			"repeat_every": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: resourceYandexBackupIntervalValidationFunc,
			},

			"monthdays": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 31),
				},
			},

			"include_last_day_of_month": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"months": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 12),
				},
			},

			"type": {
				Type:         schema.TypeString,
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
