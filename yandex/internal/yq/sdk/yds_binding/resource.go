package yds_binding

import (
	"slices"
	"strings"

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

func shouldSuppressDiffForColumnType(k, old, new string, d *schema.ResourceData) bool {
	oldLower := strings.ToLower(old)
	newLower := strings.ToLower(new)
	if oldLower == newLower {
		return true
	}

	textTypes := []string{"utf8", "text"}
	if slices.Contains(textTypes, oldLower) && slices.Contains(textTypes, newLower) {
		return true
	}

	blobTypes := []string{"string", "bytes"}
	if slices.Contains(blobTypes, oldLower) && slices.Contains(blobTypes, newLower) {
		return true
	}
	return false
}

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
		AttributeStream: {
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
						Type:             schema.TypeString,
						Description:      "Column data type. YQL data types are used.",
						Required:         true,
						ValidateFunc:     validation.NoZeroValues,
						DiffSuppressFunc: shouldSuppressDiffForColumnType,
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
