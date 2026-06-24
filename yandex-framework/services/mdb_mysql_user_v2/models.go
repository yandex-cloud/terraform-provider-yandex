package mdb_mysql_user_v2

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var rolesMap = map[string]mysql.Permission_Privilege{
	"ALL":                     mysql.Permission_ALL_PRIVILEGES,
	"ALTER":                   mysql.Permission_ALTER,
	"ALTER_ROUTINE":           mysql.Permission_ALTER_ROUTINE,
	"CREATE":                  mysql.Permission_CREATE,
	"CREATE_ROUTINE":          mysql.Permission_CREATE_ROUTINE,
	"CREATE_TEMPORARY_TABLES": mysql.Permission_CREATE_TEMPORARY_TABLES,
	"CREATE_VIEW":             mysql.Permission_CREATE_VIEW,
	"DELETE":                  mysql.Permission_DELETE,
	"DROP":                    mysql.Permission_DROP,
	"EVENT":                   mysql.Permission_EVENT,
	"EXECUTE":                 mysql.Permission_EXECUTE,
	"INDEX":                   mysql.Permission_INDEX,
	"INSERT":                  mysql.Permission_INSERT,
	"LOCK_TABLES":             mysql.Permission_LOCK_TABLES,
	"SELECT":                  mysql.Permission_SELECT,
	"SHOW_VIEW":               mysql.Permission_SHOW_VIEW,
	"TRIGGER":                 mysql.Permission_TRIGGER,
	"UPDATE":                  mysql.Permission_UPDATE,
	"REFERENCES":              mysql.Permission_REFERENCES,
}

var revertedRolesMap = map[mysql.Permission_Privilege]string{
	mysql.Permission_ALL_PRIVILEGES:          "ALL",
	mysql.Permission_ALTER:                   "ALTER",
	mysql.Permission_ALTER_ROUTINE:           "ALTER_ROUTINE",
	mysql.Permission_CREATE:                  "CREATE",
	mysql.Permission_CREATE_ROUTINE:          "CREATE_ROUTINE",
	mysql.Permission_CREATE_TEMPORARY_TABLES: "CREATE_TEMPORARY_TABLES",
	mysql.Permission_CREATE_VIEW:             "CREATE_VIEW",
	mysql.Permission_DELETE:                  "DELETE",
	mysql.Permission_DROP:                    "DROP",
	mysql.Permission_EVENT:                   "EVENT",
	mysql.Permission_EXECUTE:                 "EXECUTE",
	mysql.Permission_INDEX:                   "INDEX",
	mysql.Permission_INSERT:                  "INSERT",
	mysql.Permission_LOCK_TABLES:             "LOCK_TABLES",
	mysql.Permission_SELECT:                  "SELECT",
	mysql.Permission_SHOW_VIEW:               "SHOW_VIEW",
	mysql.Permission_TRIGGER:                 "TRIGGER",
	mysql.Permission_UPDATE:                  "UPDATE",
	mysql.Permission_REFERENCES:              "REFERENCES",
}

func bindDatabaseRole(role string) (mysql.Permission_Privilege, error) {
	upper := strings.ToUpper(role)
	if v, ok := rolesMap[upper]; ok {
		return v, nil
	}
	return mysql.Permission_PRIVILEGE_UNSPECIFIED, fmt.Errorf(
		"unknown MySQL permission role: %q, supported values: %s",
		role, supportedRoles(),
	)
}

func unbindDatabaseRole(priv mysql.Permission_Privilege) string {
	if name, ok := revertedRolesMap[priv]; ok {
		return name
	}
	return mysql.Permission_Privilege_name[int32(priv)]
}

func supportedRoles() string {
	names := make([]string, 0, len(rolesMap))
	for k := range rolesMap {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

type User struct {
	Id                   types.String   `tfsdk:"id"`
	ClusterID            types.String   `tfsdk:"cluster_id"`
	Name                 types.String   `tfsdk:"name"`
	Password             types.String   `tfsdk:"password"`
	GeneratePassword     types.Bool     `tfsdk:"generate_password"`
	Permissions          types.Set      `tfsdk:"permission"`
	GlobalPermissions    types.Set      `tfsdk:"global_permissions"`
	ConnectionLimits     types.List     `tfsdk:"connection_limits"`
	AuthenticationPlugin types.String   `tfsdk:"authentication_plugin"`
	ConnectionManager    types.Map      `tfsdk:"connection_manager"`
	DeletionProtection   types.String   `tfsdk:"deletion_protection_mode"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}

type Permission struct {
	DatabaseName types.String `tfsdk:"database_name"`
	Roles        types.List   `tfsdk:"roles"`
}

func permissionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"database_name": types.StringType,
		"roles":         types.ListType{ElemType: types.StringType},
	}
}

type ConnectionLimits struct {
	MaxQuestionsPerHour   types.Int64 `tfsdk:"max_questions_per_hour"`
	MaxUpdatesPerHour     types.Int64 `tfsdk:"max_updates_per_hour"`
	MaxConnectionsPerHour types.Int64 `tfsdk:"max_connections_per_hour"`
	MaxUserConnections    types.Int64 `tfsdk:"max_user_connections"`
}

func connectionLimitsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"max_questions_per_hour":   types.Int64Type,
		"max_updates_per_hour":     types.Int64Type,
		"max_connections_per_hour": types.Int64Type,
		"max_user_connections":     types.Int64Type,
	}
}

func specToState(ctx context.Context, spec *mysql.User, state *User, diags *diag.Diagnostics) {
	state.Id = types.StringValue(resourceid.Construct(spec.ClusterId, spec.Name))
	state.ClusterID = types.StringValue(spec.ClusterId)
	state.Name = types.StringValue(spec.Name)

	permObjType := types.ObjectType{AttrTypes: permissionAttrTypes()}
	permObjs := make([]attr.Value, 0, len(spec.Permissions))
	for _, p := range spec.Permissions {
		roles := make([]attr.Value, 0, len(p.Roles))
		for _, r := range p.Roles {
			roles = append(roles, types.StringValue(unbindDatabaseRole(r)))
		}
		rolesList, d := types.ListValue(types.StringType, roles)
		diags.Append(d...)

		obj, d := types.ObjectValue(permissionAttrTypes(), map[string]attr.Value{
			"database_name": types.StringValue(p.DatabaseName),
			"roles":         rolesList,
		})
		diags.Append(d...)
		permObjs = append(permObjs, obj)
	}
	permSet, d := types.SetValue(permObjType, permObjs)
	diags.Append(d...)
	state.Permissions = permSet

	gpObjs := make([]attr.Value, 0, len(spec.GlobalPermissions))
	for _, gp := range spec.GlobalPermissions {
		gpObjs = append(gpObjs, types.StringValue(mysql.GlobalPermission_name[int32(gp)]))
	}
	gpSet, d := types.SetValue(types.StringType, gpObjs)
	diags.Append(d...)
	state.GlobalPermissions = gpSet

	state.ConnectionLimits = flattenConnectionLimits(ctx, spec, diags)

	if spec.AuthenticationPlugin != mysql.AuthPlugin_AUTH_PLUGIN_UNSPECIFIED {
		state.AuthenticationPlugin = types.StringValue(
			mysql.AuthPlugin_name[int32(spec.AuthenticationPlugin)],
		)
	} else {
		state.AuthenticationPlugin = types.StringNull()
	}

	cmAttrs := map[string]attr.Value{}
	if spec.ConnectionManager != nil && spec.ConnectionManager.ConnectionId != "" {
		cmAttrs["connection_id"] = types.StringValue(spec.ConnectionManager.ConnectionId)
	}
	cmMap, d := types.MapValue(types.StringType, cmAttrs)
	diags.Append(d...)
	state.ConnectionManager = cmMap

	state.DeletionProtection = types.StringValue(spec.DeletionProtectionMode.String())
}

func flattenConnectionLimits(
	ctx context.Context,
	spec *mysql.User,
	diags *diag.Diagnostics,
) types.List {
	cl := spec.ConnectionLimits
	if cl == nil {
		empty, d := types.ListValue(
			types.ObjectType{AttrTypes: connectionLimitsAttrTypes()},
			[]attr.Value{},
		)
		diags.Append(d...)
		return empty
	}

	obj, d := types.ObjectValue(connectionLimitsAttrTypes(), map[string]attr.Value{
		"max_questions_per_hour":   wrapperToInt64(cl.MaxQuestionsPerHour),
		"max_updates_per_hour":     wrapperToInt64(cl.MaxUpdatesPerHour),
		"max_connections_per_hour": wrapperToInt64(cl.MaxConnectionsPerHour),
		"max_user_connections":     wrapperToInt64(cl.MaxUserConnections),
	})
	diags.Append(d...)

	lst, d := types.ListValue(
		types.ObjectType{AttrTypes: connectionLimitsAttrTypes()},
		[]attr.Value{obj},
	)
	diags.Append(d...)
	return lst
}

func wrapperToInt64(w *wrapperspb.Int64Value) types.Int64 {
	if w == nil {
		return types.Int64Value(-1)
	}
	return types.Int64Value(w.Value)
}

func stateToSpec(ctx context.Context, state *User, diags *diag.Diagnostics) *mysql.UserSpec {
	spec := &mysql.UserSpec{
		Name: state.Name.ValueString(),
	}

	if !state.Password.IsNull() && !state.Password.IsUnknown() {
		spec.Password = state.Password.ValueString()
	}

	if !state.GeneratePassword.IsNull() && !state.GeneratePassword.IsUnknown() {
		spec.GeneratePassword = wrapperspb.Bool(state.GeneratePassword.ValueBool())
	}

	if !state.Permissions.IsNull() && !state.Permissions.IsUnknown() {
		var perms []Permission
		d := state.Permissions.ElementsAs(ctx, &perms, false)
		diags.Append(d...)

		for _, p := range perms {
			var roles []string
			d := p.Roles.ElementsAs(ctx, &roles, false)
			diags.Append(d...)

			priv := make([]mysql.Permission_Privilege, 0, len(roles))
			for _, r := range roles {
				role, err := bindDatabaseRole(r)
				if err != nil {
					diags.AddError("Invalid permission role", err.Error())
					continue
				}
				priv = append(priv, role)
			}
			spec.Permissions = append(spec.Permissions, &mysql.Permission{
				DatabaseName: p.DatabaseName.ValueString(),
				Roles:        priv,
			})
		}
	}

	if !state.GlobalPermissions.IsNull() && !state.GlobalPermissions.IsUnknown() {
		var gps []string
		d := state.GlobalPermissions.ElementsAs(ctx, &gps, false)
		diags.Append(d...)

		for _, gp := range gps {
			v, ok := mysql.GlobalPermission_value[gp]
			if !ok {
				diags.AddError(
					"Invalid global permission",
					fmt.Sprintf("Unknown MySQL global permission: %q", gp),
				)
				continue
			}
			spec.GlobalPermissions = append(spec.GlobalPermissions, mysql.GlobalPermission(v))
		}
	}

	if !state.ConnectionLimits.IsNull() && !state.ConnectionLimits.IsUnknown() {
		var cls []ConnectionLimits
		d := state.ConnectionLimits.ElementsAs(ctx, &cls, false)
		diags.Append(d...)

		if len(cls) > 0 {
			cl := cls[0]
			spec.ConnectionLimits = &mysql.ConnectionLimits{
				MaxQuestionsPerHour:   int64ToWrapper(cl.MaxQuestionsPerHour),
				MaxUpdatesPerHour:     int64ToWrapper(cl.MaxUpdatesPerHour),
				MaxConnectionsPerHour: int64ToWrapper(cl.MaxConnectionsPerHour),
				MaxUserConnections:    int64ToWrapper(cl.MaxUserConnections),
			}
		}
	}

	if !state.AuthenticationPlugin.IsNull() && !state.AuthenticationPlugin.IsUnknown() {
		v, ok := mysql.AuthPlugin_value[state.AuthenticationPlugin.ValueString()]
		if !ok {
			diags.AddError(
				"Invalid authentication_plugin",
				fmt.Sprintf("Unknown authentication plugin: %q", state.AuthenticationPlugin.ValueString()),
			)
		} else {
			spec.AuthenticationPlugin = mysql.AuthPlugin(v)
		}
	}

	return spec
}

func int64ToWrapper(v types.Int64) *wrapperspb.Int64Value {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	if v.ValueInt64() < 0 {
		return nil
	}
	return wrapperspb.Int64(v.ValueInt64())
}

func getDeletionProtectionModeValue(mode types.String) mysql.DeletionProtectionMode {
	if mode.IsNull() || mode.IsUnknown() {
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_DISABLED
	}
	switch mode.ValueString() {
	case "DELETION_PROTECTION_MODE_ENABLED":
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_ENABLED
	case "DELETION_PROTECTION_MODE_DISABLED":
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_DISABLED
	case "DELETION_PROTECTION_MODE_INHERITED":
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_INHERITED
	default:
		return mysql.DeletionProtectionMode_DELETION_PROTECTION_MODE_DISABLED
	}
}
