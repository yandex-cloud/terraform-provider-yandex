package mdb_clickhouse_user

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon/usersettings"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type User interface {
	SetId(id types.String)
	SetClusterId(clusterId types.String)
	SetName(name types.String)
	SetPassword(password types.String)
	SetPermissions(permissions types.Set)
	GetPermissions() types.Set
	SetSettings(settings types.Object)
	GetSettings() types.Object
	SetQuotas(quotas types.Set)
	GetQuotas() types.Set
	SetConnectionManager(connectionManager types.Object)
	GetConnectionManager() types.Object
}

type ResourceUser struct {
	Id                types.String   `tfsdk:"id"`
	ClusterID         types.String   `tfsdk:"cluster_id"`
	Name              types.String   `tfsdk:"name"`
	Password          types.String   `tfsdk:"password"`
	GeneratePassword  types.Bool     `tfsdk:"generate_password"`
	AuthMethod        types.String   `tfsdk:"auth_method"`
	Permissions       types.Set      `tfsdk:"permission"`
	Settings          types.Object   `tfsdk:"settings"`
	Quotas            types.Set      `tfsdk:"quota"`
	ConnectionManager types.Object   `tfsdk:"connection_manager"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func (ru *ResourceUser) SetId(id types.String) {
	ru.Id = id
}

func (ru *ResourceUser) SetClusterId(clusterId types.String) {
	ru.ClusterID = clusterId
}

func (ru *ResourceUser) SetName(name types.String) {
	ru.Name = name
}

func (ru *ResourceUser) SetPassword(password types.String) {
	ru.Password = password
}

func (ru *ResourceUser) SetAuthMethod(authMethod types.String) {
	ru.AuthMethod = authMethod
}

func (ru *ResourceUser) SetPermissions(permissions types.Set) {
	ru.Permissions = permissions
}

func (ru *ResourceUser) GetPermissions() types.Set {
	return ru.Permissions
}

func (ru *ResourceUser) SetSettings(settings types.Object) {
	ru.Settings = settings
}

func (ru *ResourceUser) GetSettings() types.Object {
	return ru.Settings
}

func (ru *ResourceUser) SetQuotas(quotas types.Set) {
	ru.Quotas = quotas
}

func (ru *ResourceUser) GetQuotas() types.Set {
	return ru.Quotas
}

func (ru *ResourceUser) SetConnectionManager(connectionManager types.Object) {
	ru.ConnectionManager = connectionManager
}

func (ru *ResourceUser) GetConnectionManager() types.Object {
	return ru.ConnectionManager
}

type DatasourceUser struct {
	Id                types.String `tfsdk:"id"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	Name              types.String `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	AuthMethod        types.String `tfsdk:"auth_method"`
	Permissions       types.Set    `tfsdk:"permission"`
	Settings          types.Object `tfsdk:"settings"`
	Quotas            types.Set    `tfsdk:"quota"`
	ConnectionManager types.Object `tfsdk:"connection_manager"`
}

func (du *DatasourceUser) SetId(id types.String) {
	du.Id = id
}

func (du *DatasourceUser) SetClusterId(clusterId types.String) {
	du.ClusterID = clusterId
}

func (du *DatasourceUser) SetName(name types.String) {
	du.Name = name
}

func (du *DatasourceUser) SetPassword(password types.String) {
	du.Password = password
}

func (du *DatasourceUser) SetAuthMethod(authMethod types.String) {
	du.AuthMethod = authMethod
}

func (du *DatasourceUser) SetPermissions(permissions types.Set) {
	du.Permissions = permissions
}
func (du *DatasourceUser) GetPermissions() types.Set {
	return du.Permissions
}

func (du *DatasourceUser) SetSettings(settings types.Object) {
	du.Settings = settings
}

func (du *DatasourceUser) GetSettings() types.Object {
	return du.Settings
}

func (du *DatasourceUser) SetQuotas(quotas types.Set) {
	du.Quotas = quotas
}

func (du *DatasourceUser) GetQuotas() types.Set {
	return du.Quotas
}

func (du *DatasourceUser) SetConnectionManager(connectionManager types.Object) {
	du.ConnectionManager = connectionManager
}

func (du *DatasourceUser) GetConnectionManager() types.Object {
	return du.ConnectionManager
}

type Permission struct {
	DatabaseName types.String `tfsdk:"database_name"`
}

var permissionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"database_name": types.StringType,
	},
}

type Quota struct {
	IntervalDuration types.Int64 `tfsdk:"interval_duration"`
	Queries          types.Int64 `tfsdk:"queries"`
	Errors           types.Int64 `tfsdk:"errors"`
	ResultRows       types.Int64 `tfsdk:"result_rows"`
	ReadRows         types.Int64 `tfsdk:"read_rows"`
	ExecutionTime    types.Int64 `tfsdk:"execution_time"`
}

var quotaType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"interval_duration": types.Int64Type,
		"queries":           types.Int64Type,
		"errors":            types.Int64Type,
		"result_rows":       types.Int64Type,
		"read_rows":         types.Int64Type,
		"execution_time":    types.Int64Type,
	},
}

type ConnectionManager struct {
	ConnectionId types.String `tfsdk:"connection_id"`
}

var connectionManagerType = map[string]attr.Type{
	"connection_id": types.StringType,
}

func userToState(ctx context.Context, user *clickhouse.User, state User) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[TRACE] mdb_clickhouse_user: flatten state from user: %+v\n", user)
	state.SetName(types.StringValue(user.Name))
	state.SetClusterId(types.StringValue(user.ClusterId))
	if state, ok := state.(interface{ SetAuthMethod(types.String) }); ok {
		state.SetAuthMethod(getAuthMethodName(user.AuthMethod))
	}

	state.SetPermissions(flattenPermissions(ctx, user.Permissions, &diags))
	log.Printf("[TRACE] mdb_clickhouse_user: flattened permissions: %+v\n", state.GetPermissions())
	state.SetQuotas(flattenQuotas(ctx, user.Quotas, &diags))
	log.Printf("[TRACE] mdb_clickhouse_user: flattened quotas: %+v\n", state.GetQuotas())
	settingsObj := usersettings.Flatten(ctx, user.Settings, &diags)
	// handle both empty and missing setting block with "empty" UserSetting response from API to prevent inconsistent result after apply
	if !state.GetSettings().Equal(settingsObj) && state.GetSettings().IsNull() && chcommon.IsProtoMessageEmpty(user.Settings.ProtoReflect()) {
		settingsObj = state.GetSettings()
	}
	state.SetSettings(settingsObj)
	log.Printf("[TRACE] mdb_clickhouse_user: flattened settings: %+v\n", state.GetSettings())
	state.SetConnectionManager(flattenConnectionManager(ctx, user.ConnectionManager, &diags))
	log.Printf("[TRACE] mdb_clickhouse_user: flattened connection_manager: %+v\n", state.GetConnectionManager())

	return diags
}

func userFromState(ctx context.Context, state *ResourceUser) (*clickhouse.UserSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	permissions := expandPermissionsFromState(ctx, state.Permissions, &diags)
	log.Printf("[TRACE] mdb_clickhouse_user: expanded quotas: %+v\n", permissions)
	quotas := expandQuotasFromState(ctx, state.Quotas, &diags)
	log.Printf("[TRACE] mdb_clickhouse_user: expanded quotas: %+v\n", quotas)
	settings := usersettings.Expand(ctx, state.Settings, &diags)
	log.Printf("[TRACE] mdb_clickhouse_user: expanded settings: %+v\n", settings)
	return &clickhouse.UserSpec{
		Name:             state.Name.ValueString(),
		Password:         state.Password.ValueString(),
		Permissions:      permissions,
		Quotas:           quotas,
		Settings:         settings,
		GeneratePassword: wrapperspb.Bool(state.GeneratePassword.ValueBool()),
		AuthMethod:       getAuthMethodValue(state.AuthMethod),
	}, diags
}
