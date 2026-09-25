package topic

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers/topic"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/attributes"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
)

const (
	ydbTopicCodecGZIP = "gzip"
	ydbTopicCodecRAW  = "raw"
	ydbTopicCodecZSTD = "zstd"
)

func ResourceCreateFunc(cb auth.GetAuthCallback) helpers.TerraformCRUD {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		authCreds, err := cb(ctx)
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "failed to create token for YDB request",
					Detail:   err.Error(),
				},
			}
		}
		c := &caller{
			authCreds: authCreds,
		}
		return c.resourceYDBTopicCreate(ctx, d, meta)
	}
}

func ResourceReadFunc(cb auth.GetAuthCallback) helpers.TerraformCRUD {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		authCreds, err := cb(ctx)
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "failed to create token for YDB request",
					Detail:   err.Error(),
				},
			}
		}
		c := &caller{
			authCreds: authCreds,
		}
		return c.resourceYDBTopicRead(ctx, d, meta)
	}
}

func ResourceUpdateFunc(cb auth.GetAuthCallback) helpers.TerraformCRUD {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		authCreds, err := cb(ctx)
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "failed to create token for YDB request",
					Detail:   err.Error(),
				},
			}
		}
		c := &caller{
			authCreds: authCreds,
		}
		return c.resourceYDBTopicUpdate(ctx, d, meta)
	}
}

func ResourceDeleteFunc(cb auth.GetAuthCallback) helpers.TerraformCRUD {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		authCreds, err := cb(ctx)
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "failed to create token for YDB request",
					Detail:   err.Error(),
				},
			}
		}
		c := &caller{
			authCreds: authCreds,
		}
		return c.resourceYDBTopicDelete(ctx, d, meta)
	}
}

func DataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"database_endpoint": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"stream_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"description": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"partitions_count": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"supported_codecs": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(topic.YDBTopicAllowedCodecs, false),
			},
		},
		"retention_period_ms": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  1000 * 60 * 60 * 24, // 1 day
		},
		"consumer": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"supported_codecs": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice(topic.YDBTopicAllowedCodecs, false),
						},
					},
					"starting_message_timestamp_ms": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"service_type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"important": {
						Type:     schema.TypeBool,
						Optional: true,
					},
				},
			},
		},
	}
}

func ResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		attributes.DatabaseEndpoint: {
			Type:        schema.TypeString,
			Description: "YDB database endpoint.",
			Required:    true,
			ForceNew:    true,
		},
		attributeName: {
			Type:        schema.TypeString,
			Description: "Topic name.",
			Required:    true,
			ForceNew:    true,
		},
		attributes.Description: {
			Type:        schema.TypeString,
			Description: "Topic description.",
			Optional:    true,
		},
		attributePartitionsCount: {
			Type:        schema.TypeInt,
			Description: "Number of min partitions. Default value `1`.",
			Optional:    true,
			Computed:    true,
		},
		attributeMaxPartitionsCount: {
			Type:        schema.TypeInt,
			Description: "Number of max active partitions. Default value `1`.",
			Optional:    true,
			Computed:    true,
		},
		attributeSupportedCodecs: {
			Type:        schema.TypeSet,
			Description: "Supported data encodings. Can be one of `gzip`, `raw` or `zstd`.",
			Optional:    true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(topic.YDBTopicAllowedCodecs, false),
			},
			Computed: true,
		},
		attributeRetentionPeriodHours: {
			Type:        schema.TypeInt,
			Description: "Data retention time. Default value `86400000`.",
			Optional:    true,
			Computed:    true,
		},
		attributeRetentionStorageMB: {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		attributePartitionWriteSpeedKBPS: {
			Type:        schema.TypeInt,
			Description: "Maximum allowed write speed per partition. If a write speed for a given partition exceeds this value, the write speed will be capped. Default value: `1024 (1MB)`.",
			Optional:    true,
			Computed:    true,
		},
		attributeMeteringMode: {
			Type:        schema.TypeString,
			Description: "Resource metering mode (`reserved_capacity` - based on the allocated resources or `request_units` - based on actual usage). This option applies to topics in serverless databases.",
			Optional:    true,
			Computed:    true,
		},
		attributeAutoPartitioningSettings: {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					attributeAutoPartitioningStrategy: {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ValidateFunc: validation.StringInSlice([]string{
							attributeAutoPartitioningStrategyUnspecified,
							attributeAutoPartitioningStrategyDisabled,
							attributeAutoPartitioningStrategyScaleUp,
							attributeAutoPartitioningStrategyScaleUpAndDown,
							attributeAutoPartitioningStrategyPaused,
						}, false),
						Description: "The auto partitioning strategy to use",
					},
					attributeAutoPartitioningWriteSpeedStrategy: {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								attributeStabilizationWindow: {
									Type:        schema.TypeInt,
									Optional:    true,
									Computed:    true,
									Description: "The stabilization window in seconds",
								},
								attributeUpUtilizationPercent: {
									Type:        schema.TypeInt,
									Optional:    true,
									Computed:    true,
									Description: "The up utilization percentage threshold",
								},
								attributeDownUtilizationPercent: {
									Type:        schema.TypeInt,
									Optional:    true,
									Computed:    true,
									Description: "The down utilization percentage threshold",
								},
							},
						},
					},
				},
			},
		},
		attributeConsumer: {
			Type:        schema.TypeSet,
			Description: "Topic Readers.",
			Optional:    true,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					attributeName: {
						Type:         schema.TypeString,
						Description:  "Reader's name.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					attributeSupportedCodecs: {
						Type:        schema.TypeSet,
						Description: "Supported data encodings. Can be one of `gzip`, `raw` or `zstd`.",
						Optional:    true,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice(topic.YDBTopicAllowedCodecs, false),
						},
						Computed: true,
					},
					attributeConsumerStartingMessageTimestampMS: {
						Type:        schema.TypeInt,
						Description: "Timestamp in UNIX timestamp format from which the reader will start reading data. Default value `0`.",
						Optional:    true,
						Computed:    true,
					},
					attributeConsumerImportant: {
						Type:        schema.TypeBool,
						Description: "Defines an important consumer. No data will be deleted from the topic until all the important consumers read them. Default value `false`.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}
