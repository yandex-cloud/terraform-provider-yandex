package mdbcommon

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	baseOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}
)

type BackupWindow struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}

var BackupWindowType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"hours":   types.Int64Type,
		"minutes": types.Int64Type,
	},
}

type MaintenanceWindow struct {
	Type types.String `tfsdk:"type"`
	Day  types.String `tfsdk:"day"`
	Hour types.Int64  `tfsdk:"hour"`
}

var MaintenanceWindowType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"day":  types.StringType,
		"hour": types.Int64Type,
	},
}

type Resource struct {
	ResourcePresetId types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	DiskTypeId       types.String `tfsdk:"disk_type_id"`
}

type resourceModel[T any] interface {
	SetResourcePresetId(string)
	SetDiskSize(int64)
	SetDiskTypeId(string)

	GetResourcePresetId() string
	GetDiskSize() int64
	GetDiskTypeId() string
	*T
}

type accessModel[T any] interface {
	SetDataLens(bool)
	SetDataTransfer(bool)
	SetServerless(bool)
	SetWebSql(bool)

	GetDataLens() bool
	GetDataTransfer() bool
	GetServerless() bool
	GetWebSql() bool

	*T
}

var ResourceType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"resource_preset_id": types.StringType,
		"disk_type_id":       types.StringType,
		"disk_size":          types.Int64Type,
	},
}

const (
	anytimeType = "ANYTIME"
	weeklyType  = "WEEKLY"
)

var (
	weekdayNums = map[int32]string{
		0: "WEEK_DAY_UNSPECIFIED",
		1: "MON",
		2: "TUE",
		3: "WED",
		4: "THU",
		5: "FRI",
		6: "SAT",
		7: "SUN",
	}
	weekdayNames = map[string]int32{
		"WEEK_DAY_UNSPECIFIED": 0,
		"MON":                  1,
		"TUE":                  2,
		"WED":                  3,
		"THU":                  4,
		"FRI":                  5,
		"SAT":                  6,
		"SUN":                  7,
	}
)

type weeklyMaintenanceWindow[T any, WD ~int32] interface {
	SetDay(WD)
	SetHour(int64)
	GetDay() WD
	GetHour() int64

	*T
}

type anytimeMaintenanceWindow[T any] interface {
	*T
}

type maintenanceWindow[
	T any,
	VW any, VA any,
	WD ~int32,
	W weeklyMaintenanceWindow[VW, WD],
	A anytimeMaintenanceWindow[VA],
] interface {
	SetAnytime(A)
	SetWeeklyMaintenanceWindow(W)
	GetAnytime() A
	GetWeeklyMaintenanceWindow() W
	*T
}

type Access struct {
	DataLens     types.Bool `tfsdk:"data_lens"`
	WebSql       types.Bool `tfsdk:"web_sql"`
	Serverless   types.Bool `tfsdk:"serverless"`
	DataTransfer types.Bool `tfsdk:"data_transfer"`
}

var AccessAttrTypes = map[string]attr.Type{
	"data_lens":     types.BoolType,
	"web_sql":       types.BoolType,
	"serverless":    types.BoolType,
	"data_transfer": types.BoolType,
}
