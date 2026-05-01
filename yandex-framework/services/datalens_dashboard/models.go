package datalens_dashboard

import "github.com/hashicorp/terraform-plugin-framework/types"

// dashboardModel mirrors the DataLens dashboard wire payload. The API request
// is `{entry: {workbookId, name, key, annotation, meta, data}}` and the
// response is `{entry: {entryId, workbookId, name, key, annotation, meta,
// data, createdAt, updatedAt, revId, ...}}`. The model follows that nesting
// one to one — identifiers and timestamps live inside `entry`, never at the
// top level — to keep wire mapping simple. Top-level `Id` is taken from
// `entry.entryId` of the response by unmarshalDashboardResponse so Terraform's
// import flow can write it.
type dashboardModel struct {
	Id             types.String `tfsdk:"id"              wire:"id"`
	OrganizationId types.String `tfsdk:"organization_id" wire:"-"`

	Entry *dashboardEntryModel `tfsdk:"entry" wire:"entry"`
}

// dashboardEntryModel is the `entry` block on the wire. It carries the
// typed sub-blocks (annotation/meta/data), entry-level identifiers used by
// the API (workbookId, key, name) and the computed timestamps and revision
// IDs returned on read.
type dashboardEntryModel struct {
	Name        types.String              `tfsdk:"name"         wire:"name"`
	WorkbookId  types.String              `tfsdk:"workbook_id"  wire:"workbookId,nullIfEmpty"`
	CreatedAt   types.String              `tfsdk:"created_at"   wire:"createdAt"`
	UpdatedAt   types.String              `tfsdk:"updated_at"   wire:"updatedAt"`
	RevisionId  types.String              `tfsdk:"revision_id"  wire:"revId"`
	SavedId     types.String              `tfsdk:"saved_id"     wire:"savedId,nullIfEmpty"`
	PublishedId types.String              `tfsdk:"published_id" wire:"publishedId,nullIfEmpty"`
	Annotation  *dashboardAnnotationModel `tfsdk:"annotation"   wire:"annotation"`
	Meta        *dashboardMetaModel       `tfsdk:"meta"         wire:"meta,alwaysEmit"`
	Data        *dashboardDataModel       `tfsdk:"data"         wire:"data"`
}

type dashboardAnnotationModel struct {
	Description types.String `tfsdk:"description" wire:"description"`
}

type dashboardMetaModel struct {
	Title  types.String `tfsdk:"title"  wire:"title"`
	Locale types.String `tfsdk:"locale" wire:"locale"`
}

type dashboardDataModel struct {
	Counter            types.Int64             `tfsdk:"counter"             wire:"counter"`
	Salt               types.String            `tfsdk:"salt"                wire:"salt"`
	SchemeVersion      types.Int64             `tfsdk:"scheme_version"      wire:"schemeVersion"`
	AccessDescription  types.String            `tfsdk:"access_description"  wire:"accessDescription,nullIfEmpty"`
	SupportDescription types.String            `tfsdk:"support_description" wire:"supportDescription,nullIfEmpty"`
	Settings           *dashboardSettingsModel `tfsdk:"settings"            wire:"settings"`
	Tabs               []dashboardTabModel     `tfsdk:"tabs"                wire:"tabs"`
}

type dashboardSettingsModel struct {
	AutoupdateInterval    types.Int64  `tfsdk:"autoupdate_interval"      wire:"autoupdateInterval"`
	MaxConcurrentRequests types.Int64  `tfsdk:"max_concurrent_requests"  wire:"maxConcurrentRequests"`
	SilentLoading         types.Bool   `tfsdk:"silent_loading"           wire:"silentLoading"`
	DependentSelectors    types.Bool   `tfsdk:"dependent_selectors"      wire:"dependentSelectors"`
	ExpandTOC             types.Bool   `tfsdk:"expand_toc"               wire:"expandTOC"`
	HideDashTitle         types.Bool   `tfsdk:"hide_dash_title"          wire:"hideDashTitle"`
	HideTabs              types.Bool   `tfsdk:"hide_tabs"                wire:"hideTabs"`
	LoadOnlyVisibleCharts types.Bool   `tfsdk:"load_only_visible_charts" wire:"loadOnlyVisibleCharts"`
	LoadPriority          types.String `tfsdk:"load_priority"            wire:"loadPriority"`
	GlobalParams          map[string]string `tfsdk:"global_params"            wire:"globalParams"`
}

type dashboardTabModel struct {
	Id          types.String                  `tfsdk:"id"          wire:"id"`
	Title       types.String                  `tfsdk:"title"       wire:"title"`
	Items       []dashboardTabItemModel       `tfsdk:"items"       wire:"items,alwaysEmit"`
	Layout      []dashboardTabLayoutModel     `tfsdk:"layout"      wire:"layout,alwaysEmit"`
	Connections []dashboardTabConnectionModel `tfsdk:"connections" wire:"connections,alwaysEmit"`
	Aliases     map[string][][]string         `tfsdk:"aliases"     wire:"aliases,alwaysEmit"`
}

type dashboardTabLayoutModel struct {
	I      types.String `tfsdk:"i"      wire:"i"`
	X      types.Int64  `tfsdk:"x"      wire:"x"`
	Y      types.Int64  `tfsdk:"y"      wire:"y"`
	W      types.Int64  `tfsdk:"w"      wire:"w"`
	H      types.Int64  `tfsdk:"h"      wire:"h"`
	Parent types.String `tfsdk:"parent" wire:"parent"`
}

type dashboardTabConnectionModel struct {
	From types.String `tfsdk:"from" wire:"from"`
	To   types.String `tfsdk:"to"   wire:"to"`
	Kind types.String `tfsdk:"kind" wire:"kind"`
}

// dashboardTabItemModel discriminates by `type`. Variant blocks carry
// `wire:"-"`; (un)marshalDashboard does an explicit two-pass against the
// item's `data` map dispatched on `type`.
type dashboardTabItemModel struct {
	Id        types.String                `tfsdk:"id"            wire:"id"`
	Namespace types.String                `tfsdk:"namespace"     wire:"namespace"`
	Widget    *dashboardWidgetItemModel   `tfsdk:"widget"        wire:"-"`
	GroupCtl  *dashboardGroupCtlItemModel `tfsdk:"group_control" wire:"-"`
	Text      *dashboardTextItemModel     `tfsdk:"text"          wire:"-"`
	Title     *dashboardTitleItemModel    `tfsdk:"title"         wire:"-"`
	Image     *dashboardImageItemModel    `tfsdk:"image"         wire:"-"`
}

type dashboardWidgetItemModel struct {
	HideTitle types.Bool                `tfsdk:"hide_title" wire:"hideTitle"`
	Tabs      []dashboardWidgetTabModel `tfsdk:"tabs"       wire:"tabs"`
}

type dashboardWidgetTabModel struct {
	Id          types.String `tfsdk:"id"          wire:"id"`
	Title       types.String `tfsdk:"title"       wire:"title"`
	Description types.String `tfsdk:"description" wire:"description"`
	ChartId     types.String `tfsdk:"chart_id"    wire:"chartId"`
	IsDefault   types.Bool   `tfsdk:"is_default"  wire:"isDefault"`
	AutoHeight  types.Bool   `tfsdk:"auto_height" wire:"autoHeight"`
	Params      map[string]string `tfsdk:"params"      wire:"params"`
}

type dashboardGroupCtlItemModel struct {
	AutoHeight  types.Bool                    `tfsdk:"auto_height"  wire:"autoHeight"`
	ButtonApply types.Bool                    `tfsdk:"button_apply" wire:"buttonApply"`
	ButtonReset types.Bool                    `tfsdk:"button_reset" wire:"buttonReset"`
	Group       []dashboardGroupCtlEntryModel `tfsdk:"group"        wire:"group"`
}

type dashboardGroupCtlEntryModel struct {
	Id            types.String                  `tfsdk:"id"             wire:"id"`
	Namespace     types.String                  `tfsdk:"namespace"      wire:"namespace"`
	PlacementMode types.String                  `tfsdk:"placement_mode" wire:"placementMode"`
	Defaults      map[string][]string           `tfsdk:"defaults"       wire:"defaults"`
	Source        *dashboardGroupCtlSourceModel `tfsdk:"source"         wire:"source"`
}

type dashboardGroupCtlSourceModel struct {
	AcceptableValues []dashboardGroupCtlValueModel `tfsdk:"acceptable_values" wire:"acceptableValues"`
	AccentType       types.String                  `tfsdk:"accent_type"       wire:"accentType"`
	DefaultValue     []string                      `tfsdk:"default_value"     wire:"defaultValue"`
	ElementType      types.String                  `tfsdk:"element_type"      wire:"elementType"`
	FieldName        types.String                  `tfsdk:"field_name"        wire:"fieldName"`
	Hint             types.String                  `tfsdk:"hint"              wire:"hint"`
	Multiselectable  types.Bool                    `tfsdk:"multiselectable"   wire:"multiselectable"`
	Required         types.Bool                    `tfsdk:"required"          wire:"required"`
	ShowHint         types.Bool                    `tfsdk:"show_hint"         wire:"showHint"`
	Title            types.String                  `tfsdk:"title"             wire:"title"`
}

type dashboardGroupCtlValueModel struct {
	Title types.String `tfsdk:"title" wire:"title"`
	Value types.String `tfsdk:"value" wire:"value"`
}

type dashboardTextItemModel struct {
	Text types.String `tfsdk:"text" wire:"text"`
}

type dashboardTitleItemModel struct {
	Text      types.String `tfsdk:"text"        wire:"text"`
	Size      types.String `tfsdk:"size"        wire:"size"`
	ShowInTOC types.Bool   `tfsdk:"show_in_toc" wire:"showInTOC"`
}

type dashboardImageItemModel struct {
	Src         types.String `tfsdk:"src"         wire:"src"`
	AltText     types.String `tfsdk:"alt_text"    wire:"altText"`
	Description types.String `tfsdk:"description" wire:"description"`
}
