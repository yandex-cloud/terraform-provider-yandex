package yqcommon

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

// descriptions
const (
	attributeFormatDescription        = "The data format, e.g. csv_with_names, json_as_string, json_each_row, json_list, parquet, raw, tsv_with_names."
	attributeCompressionDescription   = "The data compression algorithm, e.g. brotli, bzip2, gzip, lz4, xz, zstd."
	attributeFormatSettingDescription = "Special format setting."
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

var (
	availableBindingAttributes = map[string]schema.Attribute{
		AttributeStream: schema.StringAttribute{
			MarkdownDescription: "The stream name.",
			Required:            true,
		},
		AttributePathPattern: schema.StringAttribute{
			MarkdownDescription: "The path pattern within Object Storage's bucket.",
			Required:            true,
			//ValidateFunc:        validation.NoZeroValues,
		},
		AttributeProjection: schema.MapAttribute{
			MarkdownDescription: "Projection rules.",
			Optional:            true,
			ElementType:         types.StringType,
		},
		AttributePartitionedBy: schema.ListAttribute{
			MarkdownDescription: "The list of partitioning column names.",
			Optional:            true,
			ElementType:         types.StringType,
		},
	}
)

func NewBindingResourceSchema(additionalAttributes ...string) (map[string]schema.Attribute, map[string]schema.Block) {
	attributes := map[string]schema.Attribute{
		AttributeID: schema.StringAttribute{
			MarkdownDescription: common.ResourceDescriptions["id"],
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		AttributeName: schema.StringAttribute{
			MarkdownDescription: common.ResourceDescriptions["name"],
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		AttributeConnectionID: schema.StringAttribute{
			MarkdownDescription: "The connection identifier.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		AttributeDescription: schema.StringAttribute{
			MarkdownDescription: common.ResourceDescriptions["description"],
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		AttributeFormat: schema.StringAttribute{
			MarkdownDescription: attributeFormatDescription,
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(availableFormats...),
			},
		},
		AttributeCompression: schema.StringAttribute{
			MarkdownDescription: attributeCompressionDescription,
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(availableCompressions...),
			},
		},
		AttributeFormatSetting: schema.MapAttribute{
			MarkdownDescription: attributeFormatSettingDescription,
			Optional:            true,
			ElementType:         types.StringType,
		},
	}

	for _, a := range additionalAttributes {
		b := availableBindingAttributes[a]
		if b == nil {
			panic(fmt.Sprintf("Additional attribute %v for binding not found", b))
		}
		attributes[a] = b
	}

	blocks := make(map[string]schema.Block)
	blocks[AttributeColumn] = schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				AttributeColumnName: schema.StringAttribute{
					MarkdownDescription: "Column name.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				AttributeColumnType: schema.StringAttribute{
					MarkdownDescription: "Column data type. YQL data types are used.",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				AttributeColumnNotNull: schema.BoolAttribute{
					MarkdownDescription: "A column cannot have the NULL data type. Default: `false`.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}

	return attributes, blocks
}
