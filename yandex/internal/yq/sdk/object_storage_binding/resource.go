package object_storage_binding

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	availableFormats = []string{
		"csv_with_names",
		"json_as_string",
		"json_each_row",
		"json_list",
		"parquet",
		"raw",
		"tsv_with_names",
	}

	availableCompressions = []string{
		"brotli",
		"bzip2",
		"gzip",
		"lz4",
		"xz",
		"zstd",
	}
)

func ResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		AttributeName: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		AttributeConnectionID: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		AttributeDescription: {
			Type:     schema.TypeString,
			Optional: true,
		},
		AttributePathPattern: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		AttributeFormat: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(availableFormats, true),
		},
		AttributeCompression: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(availableCompressions, true),
		},
		AttributeFormatSetting: {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		AttributeProjection: {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		AttributePartitionedBy: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		AttributeColumn: {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					AttributeColumnName: {
						Type:         schema.TypeString,
						Description:  "Column name.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					AttributeColumnType: {
						Type:         schema.TypeString,
						Description:  "Column data type. YQL data types are used.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					AttributeColumnNotNull: {
						Type:        schema.TypeBool,
						Description: "A column cannot have the NULL data type. Default: `false`.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}
