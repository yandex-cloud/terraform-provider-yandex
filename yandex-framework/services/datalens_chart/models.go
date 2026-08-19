package datalens_chart

import "github.com/hashicorp/terraform-plugin-framework/types"

// chartModel mirrors the DataLens chart wire payload. The API wraps the chart
// content in `{template, name, workbookId, key, annotation: {description},
// data: {type, version, visualization, ..., wizard|ql fields flat,
// colorsConfig, shapesConfig, geopointsConfig}}`. The model carries the same
// shape: `Annotation` and `Data` are nested blocks; `Wizard`/`Ql` are inlined
// into `data` because that's how DataLens stores them.
type chartModel struct {
	Id             types.String `tfsdk:"id"              wire:"entryId,nullIfEmpty"`
	OrganizationId types.String `tfsdk:"organization_id" wire:"-"`
	Type           types.String `tfsdk:"type"            wire:"-"` // RPC URL discriminator; injected into body["data"]["type"] in marshalChart
	WorkbookId     types.String `tfsdk:"workbook_id"     wire:"workbookId,nullIfEmpty"`
	Name           types.String `tfsdk:"name"            wire:"name,nullIfEmpty"`

	Annotation *chartAnnotationModel `tfsdk:"annotation" wire:"annotation"`
	Data       *chartDataModel       `tfsdk:"data"       wire:"data"`

	CreatedAt   types.String `tfsdk:"created_at"   wire:"createdAt"`
	UpdatedAt   types.String `tfsdk:"updated_at"   wire:"updatedAt"`
	RevisionId  types.String `tfsdk:"revision_id"  wire:"revId"`
	SavedId     types.String `tfsdk:"saved_id"     wire:"savedId,nullIfEmpty"`
	PublishedId types.String `tfsdk:"published_id" wire:"publishedId,nullIfEmpty"`
}

type chartAnnotationModel struct {
	Description types.String `tfsdk:"description" wire:"description"`
}

// chartDataModel is the `data` sub-block. Wizard/Ql carry `wire:"-"` because
// DataLens stores their fields flat at the `data` level — marshalChart and
// unmarshalChartResponse do an explicit two-pass over the same flat data map.
type chartDataModel struct {
	Version types.String `tfsdk:"version" wire:"version"`

	Visualization *chartVisualizationModel `tfsdk:"visualization"  wire:"visualization"`
	ExtraSettings *chartExtraSettingsModel `tfsdk:"extra_settings" wire:"extraSettings"`
	Colors        []chartFieldRef          `tfsdk:"colors"         wire:"colors"`
	Labels        []chartFieldRef          `tfsdk:"labels"         wire:"labels"`
	Shapes        []chartFieldRef          `tfsdk:"shapes"         wire:"shapes"`
	Tooltips      []chartFieldRef          `tfsdk:"tooltips"       wire:"tooltips"`
	Filters       []chartFieldRef          `tfsdk:"filters"        wire:"filters"`
	Sort          []chartFieldRef          `tfsdk:"sort"           wire:"sort"`
	Hierarchies   []chartFieldRef          `tfsdk:"hierarchies"    wire:"hierarchies"`
	Segments      []chartFieldRef          `tfsdk:"segments"       wire:"segments"`
	Links         []chartLinkModel         `tfsdk:"links"          wire:"links"`
	Updates       []chartFieldRef          `tfsdk:"updates"        wire:"updates"`

	// Variant blocks — flat in `data` on the wire, two-passed by marshalChart.
	Wizard *chartWizardModel `tfsdk:"wizard" wire:"-"`
	Ql     *chartQLModel     `tfsdk:"ql"     wire:"-"`

	// Server-required configuration objects DataLens always expects in `data`.
	// Defaulted to empty maps in schema so users don't have to set them.
	ColorsConfig    map[string]string `tfsdk:"colors_config"    wire:"colorsConfig"`
	ShapesConfig    map[string]string `tfsdk:"shapes_config"    wire:"shapesConfig"`
	GeopointsConfig map[string]string `tfsdk:"geopoints_config" wire:"geopointsConfig"`
}

type chartVisualizationModel struct {
	Id           types.String            `tfsdk:"id"           wire:"id"`
	Type         types.String            `tfsdk:"type"         wire:"type"`
	Placeholders []chartPlaceholderModel `tfsdk:"placeholders" wire:"placeholders"`
}

type chartPlaceholderModel struct {
	Id       types.String      `tfsdk:"id"       wire:"id"`
	Type     types.String      `tfsdk:"type"     wire:"type"`
	Title    types.String      `tfsdk:"title"    wire:"title"`
	Required types.Bool        `tfsdk:"required" wire:"required"`
	Capacity types.Int64       `tfsdk:"capacity" wire:"capacity"`
	Items    []chartFieldRef   `tfsdk:"items"    wire:"items"`
	Settings map[string]string `tfsdk:"settings" wire:"settings"`
}

type chartFieldRef struct {
	Id              types.String `tfsdk:"id"                wire:"id"`
	Guid            types.String `tfsdk:"guid"              wire:"guid"`
	Title           types.String `tfsdk:"title"             wire:"title"`
	DatasetId       types.String `tfsdk:"dataset_id"        wire:"datasetId"`
	Type            types.String `tfsdk:"type"              wire:"type"`
	DataType        types.String `tfsdk:"data_type"         wire:"data_type"`
	InitialDataType types.String `tfsdk:"initial_data_type" wire:"initial_data_type"`
	Cast            types.String `tfsdk:"cast"              wire:"cast"`
	CalcMode        types.String `tfsdk:"calc_mode"         wire:"calc_mode"`
	Aggregation     types.String `tfsdk:"aggregation"       wire:"aggregation"`
	Source          types.String `tfsdk:"source"            wire:"source"`
	Formula         types.String `tfsdk:"formula"           wire:"formula"`
	GuidFormula     types.String `tfsdk:"guid_formula"      wire:"guid_formula"`
	Description     types.String `tfsdk:"description"       wire:"description"`
	Hidden          types.Bool   `tfsdk:"hidden"            wire:"hidden"`
	ManagedBy       types.String `tfsdk:"managed_by"        wire:"managed_by"`
	AvatarId        types.String `tfsdk:"avatar_id"         wire:"avatar_id"`
	UiSettings      types.String `tfsdk:"ui_settings"       wire:"ui_settings"`
}

type chartLinkModel struct {
	Id     types.String          `tfsdk:"id"     wire:"id"`
	Fields []chartLinkFieldModel `tfsdk:"fields" wire:"fields"`
}

type chartLinkFieldModel struct {
	DatasetId types.String      `tfsdk:"dataset_id" wire:"datasetId"`
	Field     map[string]string `tfsdk:"field"      wire:"field"`
}

type chartExtraSettingsModel struct {
	Title              types.String                 `tfsdk:"title"                wire:"title"`
	TitleMode          types.String                 `tfsdk:"title_mode"           wire:"titleMode"`
	IndicatorTitleMode types.String                 `tfsdk:"indicator_title_mode" wire:"indicatorTitleMode"`
	LegendMode         types.String                 `tfsdk:"legend_mode"          wire:"legendMode"`
	PivotInlineSort    types.String                 `tfsdk:"pivot_inline_sort"    wire:"pivotInlineSort"`
	Stacking           types.String                 `tfsdk:"stacking"             wire:"stacking"`
	TooltipSum         types.String                 `tfsdk:"tooltip_sum"          wire:"tooltipSum"`
	Feed               types.String                 `tfsdk:"feed"                 wire:"feed"`
	Pagination         types.String                 `tfsdk:"pagination"           wire:"pagination"`
	Limit              types.Int64                  `tfsdk:"limit"                wire:"limit"`
	NavigatorSettings  *chartNavigatorSettingsModel `tfsdk:"navigator_settings"   wire:"navigatorSettings"`
}

type chartNavigatorSettingsModel struct {
	IsNavigatorAvailable types.Bool `tfsdk:"is_navigator_available" wire:"isNavigatorAvailable"`
	SelectedLines        []string   `tfsdk:"selected_lines"         wire:"selectedLines"`
}

type chartWizardModel struct {
	DatasetsIds           []string                          `tfsdk:"datasets_ids"            wire:"datasetsIds"`
	DatasetsPartialFields [][]chartDatasetPartialFieldModel `tfsdk:"datasets_partial_fields" wire:"datasetsPartialFields"`
	Convert               types.Bool                        `tfsdk:"convert"                 wire:"convert"`
}

type chartDatasetPartialFieldModel struct {
	Guid     types.String `tfsdk:"guid"      wire:"guid"`
	Title    types.String `tfsdk:"title"     wire:"title"`
	CalcMode types.String `tfsdk:"calc_mode" wire:"calc_mode"`
}

type chartQLModel struct {
	ChartType  types.String         `tfsdk:"chart_type"  wire:"chartType"`
	Connection *chartQLConnRefModel `tfsdk:"connection"  wire:"connection"`
	QueryValue types.String         `tfsdk:"query_value" wire:"queryValue"`
	Queries    []chartQLQueryModel  `tfsdk:"queries"     wire:"queries"`
	Params     []chartQLParamModel  `tfsdk:"params"      wire:"params"`
	Order      types.String         `tfsdk:"order"       wire:"order"`
}

type chartQLConnRefModel struct {
	EntryId types.String `tfsdk:"entry_id" wire:"entryId"`
	Type    types.String `tfsdk:"type"     wire:"type"`
}

type chartQLQueryModel struct {
	Value  types.String        `tfsdk:"value"  wire:"value"`
	Hidden types.Bool          `tfsdk:"hidden" wire:"hidden"`
	Params []chartQLParamModel `tfsdk:"params" wire:"params"`
}

type chartQLParamModel struct {
	Name         types.String `tfsdk:"name"          wire:"name"`
	Type         types.String `tfsdk:"type"          wire:"type"`
	DefaultValue types.String `tfsdk:"default_value" wire:"defaultValue"`
}
