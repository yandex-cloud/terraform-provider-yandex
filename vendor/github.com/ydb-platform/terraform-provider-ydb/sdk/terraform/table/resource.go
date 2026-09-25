package table

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
	"github.com/ydb-platform/terraform-provider-ydb/internal/resources/table"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
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

		h := table.NewHandler(authCreds)
		return h.Create(ctx, d, meta)
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

		h := table.NewHandler(authCreds)
		return h.Update(ctx, d, meta)
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

		h := table.NewHandler(authCreds)
		return h.Delete(ctx, d, meta)
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

		h := table.NewHandler(authCreds)
		return h.Read(ctx, d, meta)
	}
}

func ResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"path": {
			Type:         schema.TypeString,
			Description:  "Table path.",
			Required:     true,
			ForceNew:     true,
			ValidateFunc: helpers.YdbTablePathCheck,
		},
		"connection_string": {
			Type:        schema.TypeString,
			Description: "Connection string for database.",
			ForceNew:    true,
			Required:    true,
		},
		"column": {
			Type:        schema.TypeSet,
			Description: "A list of column configuration options.",
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:         schema.TypeString,
						Description:  "Column name.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"type": {
						Type:         schema.TypeString,
						Description:  "Column data type. YQL data types are used.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"family": {
						Type:         schema.TypeString,
						Description:  "Column group.",
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"not_null": {
						Type:        schema.TypeBool,
						Description: "A column cannot have the NULL data type. Default: `false`.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
		"family": {
			Type:        schema.TypeList,
			Description: "A list of column group configuration options. The `family` block may be used to group columns into [families](https://ydb.tech/en/docs/yql/reference/syntax/create_table#column-family) to set shared parameters for them.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:         schema.TypeString,
						Description:  "Column family name.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"data": {
						Type:         schema.TypeString,
						Description:  "Type of storage device for column data in this group (acceptable values: ssd, rot (from HDD spindle rotation)).",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"compression": {
						Type:         schema.TypeString,
						Description:  "Data codec (acceptable values: off, lz4).",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
				},
			},
		},
		"primary_key": {
			Type:        schema.TypeList,
			Description: "A list of table columns to be used as primary key.",
			Required:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues, // TODO(shmel1k@): think about validate func
			},
		},
		"store": {
			Type:         schema.TypeString,
			Description:  "Table storage type. Set to `column` for column-oriented tables. Omit for row-oriented tables (default).",
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"column"}, true),
		},
		"ttl": {
			Type:        schema.TypeSet,
			Description: "The `TTL` block supports allow you to create a special column type, [TTL column](https://ydb.tech/en/docs/concepts/ttl), whose values determine the time-to-live for rows.",
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"column_name": {
						Type:         schema.TypeString,
						Description:  "Column name for TTL.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"expire_interval": {
						Type:         schema.TypeString,
						Description:  "Interval in the ISO 8601 format.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"unit": {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						ValidateFunc: helpers.YdbTTLUnitCheck,
					},
				},
			},
		},
		"attributes": {
			Type:        schema.TypeMap,
			Description: "A map of table attributes.",
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"partitioning_settings": {
			Type:        schema.TypeList,
			Description: "Table partitioning settings.",
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"uniform_partitions": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"partition_at_keys": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"keys": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"auto_partitioning_min_partitions_count": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"auto_partitioning_max_partitions_count": {
						Type:             schema.TypeInt,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressWhenColumnStore,
					},
					"auto_partitioning_partition_size_mb": {
						Type:             schema.TypeInt,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressWhenColumnStore,
					},
					"auto_partitioning_by_load": {
						Type:             schema.TypeBool,
						Optional:         true,
						Default:          false,
						DiffSuppressFunc: suppressWhenColumnStore,
					},
					"auto_partitioning_by_size_enabled": {
						Type:             schema.TypeBool,
						Optional:         true,
						Default:          true,
						DiffSuppressFunc: suppressWhenColumnStore,
					},
					"partition_by": {
						Type:         schema.TypeList,
						Description:  "Partitioning keys constitute a subset of the table's primary keys. If not set, primary keys will be used.",
						Optional:     true,
						Computed:     true,
						ForceNew:     true,
						RequiredWith: []string{"store"},
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.NoZeroValues,
							MinItems:     1,
						},
					},
				},
			},
		},
		"key_bloom_filter": {
			Type:        schema.TypeBool,
			Description: "Use the Bloom filter for the primary key.",
			Optional:    true,
			Computed:    true,
		},
		"read_replicas_settings": {
			Type:        schema.TypeString,
			Description: "Read replication settings.",
			Optional:    true,
			Computed:    true,
		},
	}
}

// suppressWhenColumnStore suppresses diff changes on partition settings.
// From the YDB documentation:
//
//	To manage data partitioning, use the AUTO_PARTITIONING_MIN_PARTITIONS_COUNT additional parameter.
//	The system ignores other partitioning parameters for column-oriented tables.
//
// https://ydb.tech/docs/en/concepts/datamodel/table?version=v25.1#olap-tables-partitioning
func suppressWhenColumnStore(k, oldValue, newValue string, d *schema.ResourceData) bool {
	store, ok := d.Get("store").(string)

	return ok && store == "column"
}
