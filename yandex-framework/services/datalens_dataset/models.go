package datalens_dataset

import "github.com/hashicorp/terraform-plugin-framework/types"

// datasetModel is both the Terraform-Framework model and the wire DTO. The
// DataLens response is `{ id, name, key, is_favorite, workbook_id,
// dataset: { ...content } }`. Wire fills the model in two passes — once at
// the top level and once for the `dataset` sub-block.
type datasetModel struct {
	Id             types.String `tfsdk:"id"              wire:"id"`
	OrganizationId types.String `tfsdk:"organization_id" wire:"-"`
	WorkbookId     types.String `tfsdk:"workbook_id"     wire:"workbook_id"`
	DirPath        types.String `tfsdk:"dir_path"        wire:"dir_path"`
	Name           types.String `tfsdk:"name"            wire:"name"`
	CreatedVia     types.String `tfsdk:"created_via"     wire:"created_via"`
	Preview        types.Bool   `tfsdk:"preview"         wire:"preview"`
	IsFavorite     types.Bool   `tfsdk:"is_favorite"     wire:"is_favorite"`

	Dataset *datasetContentModel `tfsdk:"dataset" wire:"dataset"`
}

type datasetContentModel struct {
	Description          types.String `tfsdk:"description"             wire:"description,nullIfEmpty"`
	LoadPreviewByDefault types.Bool   `tfsdk:"load_preview_by_default" wire:"load_preview_by_default"`
	TemplateEnabled      types.Bool   `tfsdk:"template_enabled"        wire:"template_enabled"`
	DataExportForbidden  types.Bool   `tfsdk:"data_export_forbidden"   wire:"data_export_forbidden"`
	SchemaUpdateEnabled  types.Bool   `tfsdk:"schema_update_enabled"   wire:"schema_update_enabled"`
	PreviewEnabled       types.Bool   `tfsdk:"preview_enabled"         wire:"preview_enabled"`

	AvatarRelations         []avatarRelationModel         `tfsdk:"avatar_relations"          wire:"avatar_relations"`
	SourceAvatars           []sourceAvatarModel           `tfsdk:"source_avatars"            wire:"source_avatars"`
	Sources                 []dataSourceModel             `tfsdk:"sources"                   wire:"sources"`
	ResultSchema            []resultSchemaFieldModel      `tfsdk:"result_schema"             wire:"result_schema"`
	ObligatoryFilters       []obligatoryFilterModel       `tfsdk:"obligatory_filters"        wire:"obligatory_filters"`
	Rls2                    map[string][]rls2EntryModel   `tfsdk:"rls2"                      wire:"rls2"`
	CacheInvalidationSource *cacheInvalidationSourceModel `tfsdk:"cache_invalidation_source" wire:"cache_invalidation_source"`
}

type cacheInvalidationSourceModel struct {
	Mode    types.String   `tfsdk:"mode"    wire:"mode"`
	Field   types.String   `tfsdk:"field"   wire:"field"`
	Sql     types.String   `tfsdk:"sql"     wire:"sql"`
	Filters []string       `tfsdk:"filters" wire:"filters"`
}

type avatarRelationModel struct {
	Id            types.String     `tfsdk:"id"              wire:"id"`
	LeftAvatarId  types.String     `tfsdk:"left_avatar_id"  wire:"left_avatar_id"`
	RightAvatarId types.String     `tfsdk:"right_avatar_id" wire:"right_avatar_id"`
	JoinType      types.String     `tfsdk:"join_type"       wire:"join_type"`
	ManagedBy     types.String     `tfsdk:"managed_by"      wire:"managed_by"`
	Required      types.Bool       `tfsdk:"required"        wire:"required"`
	Virtual       types.Bool       `tfsdk:"virtual"         wire:"virtual"`
	Conditions    []conditionModel `tfsdk:"conditions"      wire:"conditions"`
}

type conditionModel struct {
	Type     types.String   `tfsdk:"type"     wire:"type"`
	Operator types.String   `tfsdk:"operator" wire:"operator"`
	Left     *joinPartModel `tfsdk:"left"     wire:"left"`
	Right    *joinPartModel `tfsdk:"right"    wire:"right"`
}

type joinPartModel struct {
	CalcMode types.String `tfsdk:"calc_mode" wire:"calc_mode"`
	Source   types.String `tfsdk:"source"    wire:"source"`
	FieldId  types.String `tfsdk:"field_id"  wire:"field_id"`
	Formula  types.String `tfsdk:"formula"   wire:"formula"`
}

type sourceAvatarModel struct {
	Id        types.String `tfsdk:"id"         wire:"id"`
	SourceId  types.String `tfsdk:"source_id"  wire:"source_id"`
	Title     types.String `tfsdk:"title"      wire:"title"`
	IsRoot    types.Bool   `tfsdk:"is_root"    wire:"is_root"`
	ManagedBy types.String `tfsdk:"managed_by" wire:"managed_by"`
	Valid     types.Bool   `tfsdk:"valid"      wire:"valid"`
	Virtual   types.Bool   `tfsdk:"virtual"    wire:"virtual"`
}

type dataSourceModel struct {
	Id           types.String           `tfsdk:"id"            wire:"id"`
	Title        types.String           `tfsdk:"title"         wire:"title"`
	SourceType   types.String           `tfsdk:"source_type"   wire:"source_type"`
	ConnectionId types.String           `tfsdk:"connection_id" wire:"connection_id"`
	ManagedBy    types.String           `tfsdk:"managed_by"    wire:"managed_by"`
	Valid        types.Bool             `tfsdk:"valid"         wire:"valid"`
	RefSourceId  types.String           `tfsdk:"ref_source_id" wire:"ref_source_id"`
	IsRef        types.Bool             `tfsdk:"is_ref"        wire:"is_ref"`
	Parameters   *sourceParametersModel `tfsdk:"parameters"    wire:"parameters"`
	RawSchema    []rawSchemaFieldModel  `tfsdk:"raw_schema"    wire:"raw_schema"`
}

type sourceParametersModel struct {
	TableName  types.String `tfsdk:"table_name"  wire:"table_name"`
	SchemaName types.String `tfsdk:"schema_name" wire:"schema_name"`
	DbName     types.String `tfsdk:"db_name"     wire:"db_name"`
	Subsql     types.String `tfsdk:"subsql"      wire:"subsql"`
}

type rawSchemaFieldModel struct {
	Name               types.String         `tfsdk:"name"                 wire:"name"`
	Title              types.String         `tfsdk:"title"                wire:"title"`
	UserType           types.String         `tfsdk:"user_type"            wire:"user_type"`
	NativeType         *rawSchemaNativeType `tfsdk:"native_type"          wire:"native_type"`
	Description        types.String         `tfsdk:"description"          wire:"description"`
	Nullable           types.Bool           `tfsdk:"nullable"             wire:"nullable"`
	HasAutoAggregation types.Bool           `tfsdk:"has_auto_aggregation" wire:"has_auto_aggregation"`
	LockAggregation    types.Bool           `tfsdk:"lock_aggregation"     wire:"lock_aggregation"`
}

type rawSchemaNativeType struct {
	Name                types.String `tfsdk:"name"                   wire:"name"`
	Nullable            types.Bool   `tfsdk:"nullable"               wire:"nullable"`
	NativeTypeClassName types.String `tfsdk:"native_type_class_name" wire:"native_type_class_name"`
}

type resultSchemaFieldModel struct {
	Guid               types.String          `tfsdk:"guid"                 wire:"guid"`
	Title              types.String          `tfsdk:"title"                wire:"title"`
	Source             types.String          `tfsdk:"source"               wire:"source"`
	DataType           types.String          `tfsdk:"data_type"            wire:"data_type"`
	Cast               types.String          `tfsdk:"cast"                 wire:"cast"`
	Type               types.String          `tfsdk:"type"                 wire:"type"`
	Aggregation        types.String          `tfsdk:"aggregation"          wire:"aggregation"`
	CalcMode           types.String          `tfsdk:"calc_mode"            wire:"calc_mode"`
	Formula            types.String          `tfsdk:"formula"              wire:"formula"`
	GuidFormula        types.String          `tfsdk:"guid_formula"         wire:"guid_formula"`
	DefaultValue       types.String          `tfsdk:"default_value"        wire:"default_value"`
	Description        types.String          `tfsdk:"description"          wire:"description"`
	Hidden             types.Bool            `tfsdk:"hidden"               wire:"hidden"`
	ManagedBy          types.String          `tfsdk:"managed_by"           wire:"managed_by"`
	Valid              types.Bool            `tfsdk:"valid"                wire:"valid"`
	AvatarId           types.String          `tfsdk:"avatar_id"            wire:"avatar_id"`
	HasAutoAggregation types.Bool            `tfsdk:"has_auto_aggregation" wire:"has_auto_aggregation"`
	LockAggregation    types.Bool            `tfsdk:"lock_aggregation"     wire:"lock_aggregation"`
	Autoaggregated     types.Bool            `tfsdk:"autoaggregated"       wire:"autoaggregated"`
	AggregationLocked  types.Bool            `tfsdk:"aggregation_locked"   wire:"aggregation_locked"`
	ValueConstraint    *valueConstraintModel `tfsdk:"value_constraint"     wire:"value_constraint"`
}

type valueConstraintModel struct {
	Type    types.String `tfsdk:"type"    wire:"type"`
	Pattern types.String `tfsdk:"pattern" wire:"pattern"`
}

type obligatoryFilterModel struct {
	Id             types.String                 `tfsdk:"id"              wire:"id"`
	FieldGuid      types.String                 `tfsdk:"field_guid"      wire:"field_guid"`
	ManagedBy      types.String                 `tfsdk:"managed_by"      wire:"managed_by"`
	Valid          types.Bool                   `tfsdk:"valid"           wire:"valid"`
	DefaultFilters []obligatoryFilterWhereModel `tfsdk:"default_filters" wire:"default_filters"`
}

type obligatoryFilterWhereModel struct {
	Operation types.String `tfsdk:"operation" wire:"operation"`
	Values    []string     `tfsdk:"values"    wire:"values"`
	Disabled  types.Bool   `tfsdk:"disabled"  wire:"disabled"`
}

type rls2EntryModel struct {
	AllowedValue types.String      `tfsdk:"allowed_value" wire:"allowed_value"`
	PatternType  types.String      `tfsdk:"pattern_type"  wire:"pattern_type"`
	Subject      *rls2SubjectModel `tfsdk:"subject"       wire:"subject"`
}

type rls2SubjectModel struct {
	SubjectId   types.String `tfsdk:"subject_id"   wire:"subject_id"`
	SubjectType types.String `tfsdk:"subject_type" wire:"subject_type"`
}

