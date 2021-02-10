package yandex

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"

	"github.com/golang/protobuf/ptypes/wrappers"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
)

type MySQLHostSpec struct {
	HostSpec        *mysql.HostSpec
	Fqdn            string
	HasComputedFqdn bool
}

func parseMysqlEnv(e string) (mysql.Cluster_Environment, error) {
	v, ok := mysql.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(mysql.Cluster_Environment_value)), e)
	}
	return mysql.Cluster_Environment(v), nil
}

func expandMysqlDatabases(d *schema.ResourceData) ([]*mysql.DatabaseSpec, error) {
	var result []*mysql.DatabaseSpec
	dbs := d.Get("database").(*schema.Set).List()

	for _, d := range dbs {
		m := d.(map[string]interface{})
		db := &mysql.DatabaseSpec{}

		if v, ok := m["name"]; ok {
			db.Name = v.(string)
		}

		result = append(result, db)
	}
	return result, nil
}

func expandMySQLUsers(users []*mysql.User, d *schema.ResourceData) ([]*mysql.UserSpec, error) {
	usersMap := make(map[string]*mysql.User)

	for _, u := range users {
		usersMap[u.Name] = u
	}

	result := []*mysql.UserSpec{}
	usersData := d.Get("user").([]interface{})

	for _, u := range usersData {
		m := u.(map[string]interface{})
		user, _, err := expandMysqlUser(m, usersMap[(m["name"]).(string)])
		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	return result, nil
}

func expandMysqlUser(u map[string]interface{}, existsUser *mysql.User) (user *mysql.UserSpec, isDiff bool, err error) {
	user = &mysql.UserSpec{}

	if existsUser != nil {
		user.AuthenticationPlugin = existsUser.AuthenticationPlugin
		user.GlobalPermissions = existsUser.GlobalPermissions
		user.ConnectionLimits = existsUser.ConnectionLimits
		user.Permissions = existsUser.Permissions
	}

	if v, ok := u["name"]; ok {
		user.Name = v.(string)
	}

	if v, ok := u["password"]; ok {
		user.Password = v.(string)
	}

	if v, ok := u["permission"]; ok {
		a, err := expandMysqlUserPermissions(v.(*schema.Set))
		if err != nil {
			return nil, false, err
		}

		isDiff = fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", a)

		user.Permissions = a
	}

	if v, ok := u["authentication_plugin"]; ok && (v.(string) != "") {
		authenticationPlugin, ok := mysql.AuthPlugin_value[v.(string)]
		if ok {
			isDiff = isDiff || user.AuthenticationPlugin != mysql.AuthPlugin(authenticationPlugin)
			user.AuthenticationPlugin = mysql.AuthPlugin(authenticationPlugin)
		} else {
			return nil, false, fmt.Errorf("User authentication_plugin not found %v", v.(string))
		}
	}

	if v, ok := u["global_permissions"]; ok && v != nil {
		list := v.([]interface{})
		if len(list) != 0 {
			gPermission, err := expandMysqlUserGlobalPermissions(list)
			if err != nil {
				return nil, false, err
			}
			isDiff = isDiff || fmt.Sprintf("%v", user.GlobalPermissions) != fmt.Sprintf("%v", gPermission)
			user.GlobalPermissions = gPermission
		} else {
			gPermission := []mysql.GlobalPermission{}
			isDiff = isDiff || fmt.Sprintf("%v", user.GlobalPermissions) != fmt.Sprintf("%v", gPermission)
			user.GlobalPermissions = gPermission
		}
	}

	if conLimits, ok := u["connection_limits"]; ok && len(conLimits.([]interface{})) != 0 {
		conLimitMap := (conLimits.([]interface{}))[0].(map[string]interface{})

		if user.ConnectionLimits == nil {
			user.ConnectionLimits = &mysql.ConnectionLimits{}
		}

		if v, ok := conLimitMap["max_questions_per_hour"]; ok && (v.(int)) > -1 {
			isDiff = isDiff || user.ConnectionLimits.MaxQuestionsPerHour == nil || user.ConnectionLimits.MaxQuestionsPerHour.GetValue() != int64(v.(int))
			user.ConnectionLimits.MaxQuestionsPerHour = &wrappers.Int64Value{Value: int64(v.(int))}
		}
		if v, ok := conLimitMap["max_updates_per_hour"]; ok && (v.(int)) > -1 {
			isDiff = isDiff || user.ConnectionLimits.MaxUpdatesPerHour == nil || user.ConnectionLimits.MaxUpdatesPerHour.GetValue() != int64(v.(int))
			user.ConnectionLimits.MaxUpdatesPerHour = &wrappers.Int64Value{Value: int64(v.(int))}
		}
		if v, ok := conLimitMap["max_connections_per_hour"]; ok && (v.(int)) > -1 {
			isDiff = isDiff || user.ConnectionLimits.MaxConnectionsPerHour == nil || user.ConnectionLimits.MaxConnectionsPerHour.GetValue() != int64(v.(int))
			user.ConnectionLimits.MaxConnectionsPerHour = &wrappers.Int64Value{Value: int64(v.(int))}
		}
		if v, ok := conLimitMap["max_user_connections"]; ok && (v.(int)) > -1 {
			isDiff = isDiff || user.ConnectionLimits.MaxUserConnections == nil || user.ConnectionLimits.MaxUserConnections.GetValue() != int64(v.(int))
			user.ConnectionLimits.MaxUserConnections = &wrappers.Int64Value{Value: int64(v.(int))}
		}
	}

	return user, isDiff, nil
}

func expandMysqlUserGlobalPermissions(ps []interface{}) ([]mysql.GlobalPermission, error) {
	result := []mysql.GlobalPermission{}

	for _, p := range ps {
		gPermition, ok := mysql.GlobalPermission_value[p.(string)]

		if !ok {
			return nil, fmt.Errorf("User global_permissions not found %v", p.(string))
		}

		if ok && gPermition > 0 {
			result = append(result, mysql.GlobalPermission(gPermition))
		}
	}
	return result, nil
}

func expandMysqlUserPermissions(ps *schema.Set) ([]*mysql.Permission, error) {
	result := []*mysql.Permission{}

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		permission := &mysql.Permission{}
		if v, ok := m["database_name"]; ok {
			permission.DatabaseName = v.(string)
		}
		if v, ok := m["roles"]; ok {
			strings := v.([]interface{})
			stringRoles := make([]string, 0)
			for _, role := range strings {
				stringRoles = append(stringRoles, role.(string))
			}
			a, err := bindDatabaseRoles(stringRoles)
			permission.Roles = a
			if err != nil {
				return nil, err
			}
		}
		result = append(result, permission)
	}
	return result, nil
}

func expandMysqlHosts(d *schema.ResourceData) ([]*MySQLHostSpec, error) {
	var result []*MySQLHostSpec
	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host, err := expandMysqlHost(config)
		if err != nil {
			return nil, err
		}
		result = append(result, host)
	}

	return result, nil
}

func expandMysqlHost(config map[string]interface{}) (*MySQLHostSpec, error) {
	hostSpec := &mysql.HostSpec{}
	host := &MySQLHostSpec{HostSpec: hostSpec, HasComputedFqdn: false}
	if v, ok := config["zone"]; ok {
		host.HostSpec.ZoneId = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		host.HostSpec.SubnetId = v.(string)
	}

	if v, ok := config["assign_public_ip"]; ok {
		host.HostSpec.AssignPublicIp = v.(bool)
	}

	if v, ok := config["fqdn"]; ok && v.(string) != "" {
		host.HasComputedFqdn = true
		host.Fqdn = v.(string)
	}

	return host, nil
}

func expandMysqlResources(d *schema.ResourceData) *mysql.Resources {
	rs := &mysql.Resources{}

	if v, ok := d.GetOk("resources.0.resource_preset_id"); ok {
		rs.ResourcePresetId = v.(string)
	}

	if v, ok := d.GetOk("resources.0.disk_size"); ok {
		rs.DiskSize = toBytes(v.(int))
	}

	if v, ok := d.GetOk("resources.0.disk_type_id"); ok {
		rs.DiskTypeId = v.(string)
	}

	return rs
}

func mysqlDatabaseHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func mysqlUserPermissionHash(v interface{}) int {
	buf := bytes.Buffer{}
	m := v.(map[string]interface{})

	if n, ok := m["database_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", n.(string)))
	}
	if n, ok := m["roles"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", n))
	}
	return hashcode.String(buf.String())
}

func flattenMysqlResources(r *mysql.Resources) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_type_id"] = r.DiskTypeId
	res["disk_size"] = toGigabytes(r.DiskSize)

	return []map[string]interface{}{res}, nil
}

func flattenMysqlBackupWindowStart(t *timeofday.TimeOfDay) ([]interface{}, error) {
	out := map[string]interface{}{}

	out["hours"] = int(t.Hours)
	out["minutes"] = int(t.Minutes)

	return []interface{}{out}, nil
}

func expandMysqlBackupWindowStart(d *schema.ResourceData) *timeofday.TimeOfDay {
	out := &timeofday.TimeOfDay{}

	if v, ok := d.GetOk("backup_window_start.0.hours"); ok {
		out.Hours = int32(v.(int))
	}

	if v, ok := d.GetOk("backup_window_start.0.minutes"); ok {
		out.Minutes = int32(v.(int))
	}

	return out
}

func listMysqlHosts(ctx context.Context, config *Config, id string) ([]*mysql.Host, error) {
	hosts := []*mysql.Host{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().MySQL().Cluster().ListHosts(ctx, &mysql.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", id, err)
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts, nil
}

func sortMysqlHosts(hosts []*mysql.Host, specs []*MySQLHostSpec) {
	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if h.HostSpec.ZoneId == hosts[j].ZoneId {
				hosts[i], hosts[j] = hosts[j], hosts[i]
				break
			}
		}
	}
}

func flattenMysqlHosts(hs []*mysql.Host) ([]map[string]interface{}, error) {
	out := []map[string]interface{}{}

	for _, h := range hs {
		m := map[string]interface{}{}
		m["zone"] = h.ZoneId
		m["subnet_id"] = h.SubnetId
		m["assign_public_ip"] = h.AssignPublicIp
		m["fqdn"] = h.Name

		out = append(out, m)
	}

	return out, nil
}

func mysqlUsersPasswords(users []*mysql.UserSpec) map[string]string {
	result := map[string]string{}
	for _, u := range users {
		result[u.Name] = u.Password
	}
	return result
}

func listMysqlUsers(ctx context.Context, config *Config, id string) ([]*mysql.User, error) {
	users := []*mysql.User{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().MySQL().User().List(ctx, &mysql.ListUsersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of users for '%s': %s", id, err)
		}
		users = append(users, resp.Users...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return users, nil
}

func flattenMysqlUsers(us []*mysql.User, passwords map[string]string) ([]interface{}, error) {
	out := make([]interface{}, 0)

	for _, u := range us {
		ou, err := flattenMysqlUser(u)
		if err != nil {
			return nil, err
		}

		if v, ok := passwords[u.Name]; ok {
			ou["password"] = v
		}

		out = append(out, ou)
	}

	return out, nil
}

func flattenMysqlUser(u *mysql.User) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	m["name"] = u.Name

	permissions, err := flattenMysqlUserPermissions(u.Permissions)
	if err != nil {
		return nil, err
	}
	m["permission"] = permissions

	connectionLimits := flattenMysqlUserConnectionLimits(u)

	if connectionLimits != nil {
		m["connection_limits"] = connectionLimits
	}

	m["global_permissions"] = unbindGlobalPermissions(u.GlobalPermissions)

	if u.AuthenticationPlugin != 0 {
		m["authentication_plugin"] = mysql.AuthPlugin_name[int32(u.AuthenticationPlugin)]
	}

	return m, nil
}

func flattenMysqlUserPermissions(ps []*mysql.Permission) (*schema.Set, error) {
	out := schema.NewSet(mysqlUserPermissionHash, nil)

	for _, p := range ps {
		roles := unbindDatabaseRoles(p.Roles)
		op := map[string]interface{}{
			"database_name": p.DatabaseName,
			"roles":         roles,
		}

		out.Add(op)
	}

	return out, nil
}

func flattenMysqlUserConnectionLimits(u *mysql.User) []map[string]interface{} {

	if u.ConnectionLimits == nil {
		return nil
	}

	m := map[string]interface{}{}
	if u.ConnectionLimits.MaxQuestionsPerHour != nil {
		m["max_questions_per_hour"] = u.ConnectionLimits.MaxQuestionsPerHour.Value
	} else {
		m["max_questions_per_hour"] = -1
	}
	if u.ConnectionLimits.MaxUpdatesPerHour != nil {
		m["max_updates_per_hour"] = u.ConnectionLimits.MaxUpdatesPerHour.Value
	} else {
		m["max_updates_per_hour"] = -1
	}
	if u.ConnectionLimits.MaxConnectionsPerHour != nil {
		m["max_connections_per_hour"] = u.ConnectionLimits.MaxConnectionsPerHour.Value
	} else {
		m["max_connections_per_hour"] = -1
	}
	if u.ConnectionLimits.MaxUserConnections != nil {
		m["max_user_connections"] = u.ConnectionLimits.MaxUserConnections.Value
	} else {
		m["max_user_connections"] = -1
	}

	return []map[string]interface{}{m}
}

func listMysqlDatabases(ctx context.Context, config *Config, id string) ([]*mysql.Database, error) {
	databases := []*mysql.Database{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().MySQL().Database().List(ctx, &mysql.ListDatabasesRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of databases for '%s': %s", id, err)
		}

		databases = append(databases, resp.Databases...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return databases, nil
}

func flattenMysqlDatabases(dbs []*mysql.Database) *schema.Set {
	out := schema.NewSet(mysqlDatabaseHash, nil)

	for _, d := range dbs {
		m := make(map[string]interface{})
		m["name"] = d.Name
		out.Add(m)
	}

	return out
}

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
}

func getRoleNames() string {
	values := []string{}
	for k := range rolesMap {
		values = append(values, k)
	}
	sort.Strings(values)

	return strings.Join(values, ",")
}

func getRole(s string) (mysql.Permission_Privilege, error) {
	sup := strings.ToUpper(s)
	if role, ok := rolesMap[sup]; ok {
		return role, nil
	}

	return mysql.Permission_PRIVILEGE_UNSPECIFIED, fmt.Errorf("unsupported database permission role flag: %v, supported values: %v", s, getRoleNames())
}

func bindDatabaseRoles(permissions []string) ([]mysql.Permission_Privilege, error) {
	var roles []mysql.Permission_Privilege
	for _, v := range permissions {
		role, err := getRole(v)

		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func unbindDatabaseRoles(permissions []mysql.Permission_Privilege) []string {
	var roles []string
	for _, v := range permissions {
		role := revertedRolesMap[v]
		roles = append(roles, role)
	}

	return roles
}

func unbindGlobalPermissions(globalPermissions []mysql.GlobalPermission) []string {
	var roles []string
	for _, v := range globalPermissions {
		roles = append(roles, mysql.GlobalPermission_name[int32(v)])
	}

	return roles
}

// parseStringToTime parse string to time, when s is 0 or is "" then now time format (unix second or "2006-01-02T15:04:05" )
func parseStringToTime(s string) (t time.Time, err error) {
	if s == "" {
		return time.Now(), nil
	}
	if s == "0" {
		return time.Now(), nil
	}

	if timeInt, err := strconv.Atoi(s); err == nil {
		return time.Unix(int64(timeInt), 0), nil
	}

	return time.Parse("2006-01-02T15:04:05", s)

}

func stringToTimeValidateFunc(value interface{}, key string) (fields []string, errors []error) {
	if strTime, ok := value.(string); ok {
		_, err := parseStringToTime(strTime)
		if err != nil {
			errors = append(errors, err)
		}
	} else {
		errors = append(errors, fmt.Errorf("value %v is not string", value))
	}

	return fields, errors
}

func flattenMySQLAccess(a *mysql.Access) ([]interface{}, error) {
	if a == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["data_lens"] = a.DataLens
	out["web_sql"] = a.WebSql

	return []interface{}{out}, nil
}

func expandMySQLAccess(d *schema.ResourceData) *mysql.Access {
	if _, ok := d.GetOkExists("access"); !ok {
		return nil
	}

	out := &mysql.Access{}

	if v, ok := d.GetOk("access.0.data_lens"); ok {
		out.DataLens = v.(bool)
	}

	if v, ok := d.GetOk("access.0.web_sql"); ok {
		out.WebSql = v.(bool)
	}

	return out
}

func flattenMySQLSettingsSQLMode57(settings map[string]string, mySQLConfig *config.MysqlConfig5_7) (map[string]string, error) {
	modes := make([]int32, 0)
	for _, v := range mySQLConfig.SqlMode {
		modes = append(modes, int32(v))
	}

	return flattenMySQLSettingsSQLMode(settings, modes)
}

func flattenMySQLSettingsSQLMode80(settings map[string]string, mySQLConfig *config.MysqlConfig8_0) (map[string]string, error) {

	modes := make([]int32, 0)
	for _, v := range mySQLConfig.SqlMode {
		modes = append(modes, int32(v))
	}

	return flattenMySQLSettingsSQLMode(settings, modes)
}

func flattenMySQLSettingsSQLMode(settings map[string]string, modes []int32) (map[string]string, error) {

	sdlMode, err := mdbMySQLSettingsFieldsInfo.intSliceToString("sql_mode", modes)
	if err != nil {
		return nil, err
	}

	if sdlMode == "" {
		return settings, nil
	}

	if settings == nil {
		settings = make(map[string]string)
	}

	settings["sql_mode"] = sdlMode

	return settings, nil
}

func flattenMySQLSettings(c *mysql.ClusterConfig) (map[string]string, error) {

	if cf, ok := c.MysqlConfig.(*mysql.ClusterConfig_MysqlConfig_8_0); ok {

		settings, err := flattenResourceGenerateMapS(cf.MysqlConfig_8_0.UserConfig, false, mdbMySQLSettingsFieldsInfo, false, true)
		if err != nil {
			return nil, err
		}

		settings, err = flattenMySQLSettingsSQLMode80(settings, cf.MysqlConfig_8_0.EffectiveConfig)
		if err != nil {
			return nil, err
		}

		return settings, err
	}
	if cf, ok := c.MysqlConfig.(*mysql.ClusterConfig_MysqlConfig_5_7); ok {
		settings, err := flattenResourceGenerateMapS(cf.MysqlConfig_5_7.UserConfig, false, mdbMySQLSettingsFieldsInfo, false, true)
		if err != nil {
			return nil, err
		}

		settings, err = flattenMySQLSettingsSQLMode57(settings, cf.MysqlConfig_5_7.EffectiveConfig)
		if err != nil {
			return nil, err
		}

		return settings, err
	}

	return nil, nil
}

func expandMySQLConfigSpecSettings(d *schema.ResourceData, configSpec *mysql.ConfigSpec) (updateFieldConfigName string, err error) {

	version := configSpec.Version

	path := "mysql_config"

	if _, ok := d.GetOkExists(path); !ok {
		return "", nil
	}

	var sdlModes []int32
	sqlMode, ok := d.GetOkExists(path + ".sql_mode")
	if ok {
		sdlModes, err = mdbMySQLSettingsFieldsInfo.stringToIntSlice("sql_mode", sqlMode.(string))
		if err != nil {
			return "", err
		}
	}

	var mySQLConfig interface{}
	if version == "5.7" {
		cfg := &mysql.ConfigSpec_MysqlConfig_5_7{
			MysqlConfig_5_7: &config.MysqlConfig5_7{},
		}
		if len(sdlModes) > 0 {
			for _, v := range sdlModes {
				cfg.MysqlConfig_5_7.SqlMode = append(cfg.MysqlConfig_5_7.SqlMode, config.MysqlConfig5_7_SQLMode(v))
			}
		}
		mySQLConfig = cfg.MysqlConfig_5_7
		configSpec.MysqlConfig = cfg
		updateFieldConfigName = "mysql_config_5_7"
	} else if version == "8.0" {
		cfg := &mysql.ConfigSpec_MysqlConfig_8_0{
			MysqlConfig_8_0: &config.MysqlConfig8_0{},
		}
		if len(sdlModes) > 0 {
			for _, v := range sdlModes {
				cfg.MysqlConfig_8_0.SqlMode = append(cfg.MysqlConfig_8_0.SqlMode, config.MysqlConfig8_0_SQLMode(v))
			}
		}
		mySQLConfig = cfg.MysqlConfig_8_0
		configSpec.MysqlConfig = cfg
		updateFieldConfigName = "mysql_config_8_0"

	} else {
		return "", nil
	}

	err = expandResourceGenerate(mdbMySQLSettingsFieldsInfo, d, mySQLConfig, path+".", true)

	if err != nil {
		return "", err
	}

	return updateFieldConfigName, nil
}

const defaultSQLModes = "ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"

var mdbMySQLSettingsFieldsInfo = newObjectFieldsInfo().
	addType(config.MysqlConfig8_0{}).
	addType(config.MysqlConfig5_7{}).
	addEnumGeneratedNames("default_authentication_plugin", config.MysqlConfig8_0_AuthPlugin_name).
	addEnumGeneratedNames("transaction_isolation", config.MysqlConfig8_0_TransactionIsolation_name).
	addEnumGeneratedNames("binlog_row_image", config.MysqlConfig8_0_BinlogRowImage_name).
	addEnumGeneratedNames("slave_parallel_type", config.MysqlConfig8_0_SlaveParallelType_name).
	addSkipEnumGeneratedNamesList("sql_mode", config.MysqlConfig8_0_SQLMode_name, defaultSQLModes, "SQLMODE_UNSPECIFIED")
