package trino_catalog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func CatalogResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"clickhouse": schema.SingleNestedAttribute{
				Validators: []validator.Object{
					onlyOneOptionValidator("Clickhouse", "connection_manager", "on_premise"),
				},
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
					"connection_manager":    connectionManagerSchema(),
					"on_premise":            onPremiseSchema(),
				},
				Optional:            true,
				Description:         "Configuration for Clickhouse connector.",
				MarkdownDescription: "Configuration for Clickhouse connector.",
			},
			"delta_lake": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
					"file_system":           fileSystemSchema(),
					"metastore":             metastoreSchema(),
				},
				Optional:            true,
				Description:         "Configuration for DeltaLake connector.",
				MarkdownDescription: "Configuration for DeltaLake connector.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "The resource description.",
				MarkdownDescription: "The resource description.",
			},
			"hive": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
					"file_system":           fileSystemSchema(),
					"metastore":             metastoreSchema(),
				},
				Optional:            true,
				Description:         "Configuration for Hive connector.",
				MarkdownDescription: "Configuration for Hive connector.",
			},
			"hudi": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
					"file_system":           fileSystemSchema(),
					"metastore":             metastoreSchema(),
				},
				Optional:            true,
				Description:         "Configuration for Hudi connector.",
				MarkdownDescription: "Configuration for Hudi connector.",
			},
			"iceberg": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
					"file_system":           fileSystemSchema(),
					"metastore":             metastoreSchema(),
				},
				Optional:            true,
				Description:         "Configuration for Iceberg connector.",
				MarkdownDescription: "Configuration for Iceberg connector.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The resource identifier.",
				MarkdownDescription: "The resource identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Trino cluster. Provided by the client when the Catalog is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "A set of key/value label pairs which assigned to resource.",
				MarkdownDescription: "A set of key/value label pairs which assigned to resource.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The resource name.",
				MarkdownDescription: "The resource name.",
			},
			"oracle": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
					"on_premise":            onPremiseSchema(),
				},
				Optional:            true,
				Description:         "Configuration for Oracle connector.",
				MarkdownDescription: "Configuration for Oracle connector.",
			},
			"postgresql": schema.SingleNestedAttribute{
				Validators: []validator.Object{
					onlyOneOptionValidator("Postgresql", "connection_manager", "on_premise"),
				},
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
					"connection_manager":    connectionManagerSchema(),
					"on_premise":            onPremiseSchema(),
				},
				Optional:            true,
				Description:         "Configuration for Postgresql connector.",
				MarkdownDescription: "Configuration for Postgresql connector.",
			},
			"sqlserver": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
					"on_premise":            onPremiseSchema(),
				},
				Optional:            true,
				Description:         "Configuration for SQLServer connector.",
				MarkdownDescription: "Configuration for SQLServer connector.",
			},
			"tpcds": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
				},
				Optional:            true,
				Description:         "Configuration for TPCDS connector.",
				MarkdownDescription: "Configuration for TPCDS connector.",
			},
			"tpch": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"additional_properties": additionalPropertiesSchema(),
				},
				Optional:            true,
				Description:         "Configuration for TPCH connector.",
				MarkdownDescription: "Configuration for TPCH connector.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				CustomType: timeouts.Type{},
			},
		},
		Description: "Catalog for Manage Trino cluster.",
	}
}

func connectionManagerSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				Required:            true,
				Description:         "Connection ID.",
				MarkdownDescription: "Connection ID.",
			},
			"connection_properties": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Additional connection properties.",
				MarkdownDescription: "Additional connection properties.",
			},
			"database": schema.StringAttribute{
				Required:            true,
				Description:         "Database.",
				MarkdownDescription: "Database.",
			},
		},
		Optional:            true,
		Description:         "Configuration for connection manager connection.",
		MarkdownDescription: "Configuration for connection manager connection.",
	}
}

func additionalPropertiesSchema() schema.MapAttribute {
	return schema.MapAttribute{
		ElementType:         types.StringType,
		Optional:            true,
		Description:         "Additional properties.",
		MarkdownDescription: "Additional properties.",
	}
}

func onPremiseSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"connection_url": schema.StringAttribute{
				Required:            true,
				Description:         "Connection to the clickhouse.",
				MarkdownDescription: "Connection to the clickhouse.",
			},
			"password": schema.StringAttribute{
				Required:            true,
				Description:         "Password of the clickhouse user.",
				MarkdownDescription: "Password of the clickhouse user.",
			},
			"user_name": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the clickhouse user.",
				MarkdownDescription: "Name of the clickhouse user.",
			},
		},
		Optional:            true,
		Description:         "Configuration for on-premise connection.",
		MarkdownDescription: "Configuration for on-premise connection.",
	}
}

func fileSystemSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Validators: []validator.Object{
			onlyOneOptionValidator("FileSystem", "s3", "external_s3"),
		},
		Attributes: map[string]schema.Attribute{
			"external_s3": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"aws_access_key": schema.StringAttribute{
						Required:            true,
						Sensitive:           true,
						Description:         "AWS access key ID for S3 authentication.",
						MarkdownDescription: "AWS access key ID for S3 authentication.",
					},
					"aws_endpoint": schema.StringAttribute{
						Required:            true,
						Description:         "AWS S3 compatible endpoint URL.",
						MarkdownDescription: "AWS S3 compatible endpoint URL.",
					},
					"aws_region": schema.StringAttribute{
						Required:            true,
						Description:         "AWS region for S3 storage.",
						MarkdownDescription: "AWS region for S3 storage.",
					},
					"aws_secret_key": schema.StringAttribute{
						Required:            true,
						Sensitive:           true,
						Description:         "AWS secret access key for S3 authentication.",
						MarkdownDescription: "AWS secret access key for S3 authentication.",
					},
				},
				Optional:            true,
				Description:         "Describes External S3 compatible file system.",
				MarkdownDescription: "Describes External S3 compatible file system.",
			},
			"s3": schema.SingleNestedAttribute{
				Attributes:          map[string]schema.Attribute{},
				Optional:            true,
				Description:         "Describes YandexCloud native S3 file system.",
				MarkdownDescription: "Describes YandexCloud native S3 file system.",
			},
		},
		Required:            true,
		Description:         "File system configuration.",
		MarkdownDescription: "File system configuration.",
	}
}

func metastoreSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				Required:            true,
				Description:         "The resource description.",
				MarkdownDescription: "The resource description.",
			},
		},
		Required:            true,
		Description:         "Metastore configuration.",
		MarkdownDescription: "Metastore configuration.",
	}
}

type CatalogModel struct {
	Id          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	ClusterId   types.String   `tfsdk:"cluster_id"`
	Description types.String   `tfsdk:"description"`
	Labels      types.Map      `tfsdk:"labels"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`

	Clickhouse types.Object `tfsdk:"clickhouse"`
	DeltaLake  types.Object `tfsdk:"delta_lake"`
	Hive       types.Object `tfsdk:"hive"`
	Hudi       types.Object `tfsdk:"hudi"`
	Iceberg    types.Object `tfsdk:"iceberg"`
	Oracle     types.Object `tfsdk:"oracle"`
	Postgresql types.Object `tfsdk:"postgresql"`
	Sqlserver  types.Object `tfsdk:"sqlserver"`
	Tpcds      types.Object `tfsdk:"tpcds"`
	Tpch       types.Object `tfsdk:"tpch"`
}

var baseOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}

type Postgresql struct {
	AdditionalProperties types.Map    `tfsdk:"additional_properties"`
	ConnectionManager    types.Object `tfsdk:"connection_manager"`
	OnPremise            types.Object `tfsdk:"on_premise"`
}

var PostgresqlT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
		"connection_manager":    ConnectionManagerT,
		"on_premise":            OnPremiseT,
	},
}

func NewPostgresqlNull() Postgresql {
	return Postgresql{
		AdditionalProperties: types.MapNull(types.StringType),
		ConnectionManager:    types.ObjectNull(ConnectionManagerT.AttrTypes),
		OnPremise:            types.ObjectNull(OnPremiseT.AttrTypes),
	}
}

func (v *Postgresql) Equal(other *Postgresql) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	if !v.ConnectionManager.Equal(other.ConnectionManager) {
		return false
	}

	if !v.OnPremise.Equal(other.OnPremise) {
		return false
	}

	return true
}

type Hive struct {
	AdditionalProperties types.Map    `tfsdk:"additional_properties"`
	FileSystem           types.Object `tfsdk:"file_system"`
	Metastore            types.Object `tfsdk:"metastore"`
}

var HiveT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
		"file_system":           FileSystemT,
		"metastore":             MetastoreT,
	},
}

func (v *Hive) Equal(other *Hive) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	if !v.FileSystem.Equal(other.FileSystem) {
		return false
	}

	if !v.Metastore.Equal(other.Metastore) {
		return false
	}

	return true
}

func NewHiveNull() Hive {
	return Hive{
		AdditionalProperties: types.MapNull(types.StringType),
		FileSystem:           types.ObjectNull(FileSystemT.AttrTypes),
		Metastore:            types.ObjectNull(MetastoreT.AttrTypes),
	}
}

type OnPremise struct {
	ConnectionUrl types.String `tfsdk:"connection_url"`
	Password      types.String `tfsdk:"password"`
	UserName      types.String `tfsdk:"user_name"`
}

var OnPremiseT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"connection_url": types.StringType,
		"password":       types.StringType,
		"user_name":      types.StringType,
	},
}

type ConnectionManager struct {
	ConnectionId         types.String `tfsdk:"connection_id"`
	ConnectionProperties types.Map    `tfsdk:"connection_properties"`
	Database             types.String `tfsdk:"database"`
}

var ConnectionManagerT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"connection_id":         types.StringType,
		"connection_properties": types.MapType{ElemType: types.StringType},
		"database":              types.StringType,
	},
}

type Metastore struct {
	Uri types.String `tfsdk:"uri"`
}

var MetastoreT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"uri": types.StringType,
	},
}

type FileSystem struct {
	ExternalS3 types.Object `tfsdk:"external_s3"`
	S3         types.Object `tfsdk:"s3"`
}

func NewFileSystemNull() FileSystem {
	return FileSystem{
		ExternalS3: types.ObjectNull(ExternalS3T.AttributeTypes()),
		S3:         types.ObjectNull(map[string]attr.Type{}),
	}
}

var FileSystemT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"external_s3": types.ObjectType{
			AttrTypes: ExternalS3T.AttributeTypes(),
		},
		"s3": types.ObjectType{
			AttrTypes: S3T.AttributeTypes(),
		},
	},
}

type S3 struct{}

var S3T = types.ObjectType{
	AttrTypes: map[string]attr.Type{},
}

type ExternalS3 struct {
	AwsAccessKey types.String `tfsdk:"aws_access_key"`
	AwsEndpoint  types.String `tfsdk:"aws_endpoint"`
	AwsRegion    types.String `tfsdk:"aws_region"`
	AwsSecretKey types.String `tfsdk:"aws_secret_key"`
}

var ExternalS3T = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"aws_access_key": types.StringType,
		"aws_endpoint":   types.StringType,
		"aws_region":     types.StringType,
		"aws_secret_key": types.StringType,
	},
}

type Clickhouse struct {
	AdditionalProperties types.Map    `tfsdk:"additional_properties"`
	ConnectionManager    types.Object `tfsdk:"connection_manager"`
	OnPremise            types.Object `tfsdk:"on_premise"`
}

var ClickhouseT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
		"connection_manager":    ConnectionManagerT,
		"on_premise":            OnPremiseT,
	},
}

func NewClickhouseNull() Clickhouse {
	return Clickhouse{
		AdditionalProperties: types.MapNull(types.StringType),
		ConnectionManager:    types.ObjectNull(ConnectionManagerT.AttrTypes),
		OnPremise:            types.ObjectNull(OnPremiseT.AttrTypes),
	}
}

func (v *Clickhouse) Equal(other *Clickhouse) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	if !v.ConnectionManager.Equal(other.ConnectionManager) {
		return false
	}

	if !v.OnPremise.Equal(other.OnPremise) {
		return false
	}

	return true
}

type DeltaLake struct {
	AdditionalProperties types.Map    `tfsdk:"additional_properties"`
	FileSystem           types.Object `tfsdk:"file_system"`
	Metastore            types.Object `tfsdk:"metastore"`
}

var DeltaLakeT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
		"file_system":           FileSystemT,
		"metastore":             MetastoreT,
	},
}

func NewDeltaLakeNull() DeltaLake {
	return DeltaLake{
		AdditionalProperties: types.MapNull(types.StringType),
		FileSystem:           types.ObjectNull(FileSystemT.AttrTypes),
		Metastore:            types.ObjectNull(MetastoreT.AttrTypes),
	}
}

func (v *DeltaLake) Equal(other *DeltaLake) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	if !v.FileSystem.Equal(other.FileSystem) {
		return false
	}

	if !v.Metastore.Equal(other.Metastore) {
		return false
	}

	return true
}

type Iceberg struct {
	AdditionalProperties types.Map    `tfsdk:"additional_properties"`
	FileSystem           types.Object `tfsdk:"file_system"`
	Metastore            types.Object `tfsdk:"metastore"`
}

var IcebergT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
		"file_system":           FileSystemT,
		"metastore":             MetastoreT,
	},
}

func NewIcebergNull() Iceberg {
	return Iceberg{
		AdditionalProperties: types.MapNull(types.StringType),
		FileSystem:           types.ObjectNull(FileSystemT.AttrTypes),
		Metastore:            types.ObjectNull(MetastoreT.AttrTypes),
	}
}

func (v *Iceberg) Equal(other *Iceberg) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	if !v.FileSystem.Equal(other.FileSystem) {
		return false
	}

	if !v.Metastore.Equal(other.Metastore) {
		return false
	}

	return true
}

type Hudi struct {
	AdditionalProperties types.Map    `tfsdk:"additional_properties"`
	FileSystem           types.Object `tfsdk:"file_system"`
	Metastore            types.Object `tfsdk:"metastore"`
}

var HudiT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
		"file_system":           FileSystemT,
		"metastore":             MetastoreT,
	},
}

func NewHudiNull() Hudi {
	return Hudi{
		AdditionalProperties: types.MapNull(types.StringType),
		FileSystem:           types.ObjectNull(FileSystemT.AttrTypes),
		Metastore:            types.ObjectNull(MetastoreT.AttrTypes),
	}
}

func (v *Hudi) Equal(other *Hudi) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	if !v.FileSystem.Equal(other.FileSystem) {
		return false
	}

	if !v.Metastore.Equal(other.Metastore) {
		return false
	}

	return true
}

type Oracle struct {
	AdditionalProperties types.Map    `tfsdk:"additional_properties"`
	OnPremise            types.Object `tfsdk:"on_premise"`
}

var OracleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
		"on_premise":            OnPremiseT,
	},
}

func NewOracleNull() Oracle {
	return Oracle{
		AdditionalProperties: types.MapNull(types.StringType),
		OnPremise:            types.ObjectNull(OnPremiseT.AttrTypes),
	}
}

func (v *Oracle) Equal(other *Oracle) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	if !v.OnPremise.Equal(other.OnPremise) {
		return false
	}

	return true
}

type Sqlserver struct {
	AdditionalProperties types.Map    `tfsdk:"additional_properties"`
	OnPremise            types.Object `tfsdk:"on_premise"`
}

var SqlserverT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
		"on_premise":            OnPremiseT,
	},
}

func NewSqlserverNull() Sqlserver {
	return Sqlserver{
		AdditionalProperties: types.MapNull(types.StringType),
		OnPremise:            types.ObjectNull(OnPremiseT.AttrTypes),
	}
}

func (v *Sqlserver) Equal(other *Sqlserver) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	if !v.OnPremise.Equal(other.OnPremise) {
		return false
	}

	return true
}

type Tpcds struct {
	AdditionalProperties types.Map `tfsdk:"additional_properties"`
}

var TpcdsT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
	},
}

func NewTpcdsNull() Tpcds {
	return Tpcds{
		AdditionalProperties: types.MapNull(types.StringType),
	}
}

func (v *Tpcds) Equal(other *Tpcds) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	return true
}

type Tpch struct {
	AdditionalProperties types.Map `tfsdk:"additional_properties"`
}

var TpchT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
	},
}

func NewTpchNull() Tpch {
	return Tpch{
		AdditionalProperties: types.MapNull(types.StringType),
	}
}

func (v *Tpch) Equal(other *Tpch) bool {
	if (v == nil && other != nil) || (v != nil && other == nil) {
		return false
	}

	if !v.AdditionalProperties.Equal(other.AdditionalProperties) {
		return false
	}

	return true
}
