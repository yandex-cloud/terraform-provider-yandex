package yandex

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/sqlserver/v1"

	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/sqlserver/v1/config"
)

func parseSQLServerEnv(e string) (sqlserver.Cluster_Environment, error) {
	v, ok := sqlserver.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(sqlserver.Cluster_Environment_value)), e)
	}
	return sqlserver.Cluster_Environment(v), nil
}

func expandSQLServerResources(d *schema.ResourceData) *sqlserver.Resources {
	rs := &sqlserver.Resources{}

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

func flattenSQLServerResources(r *sqlserver.Resources) []map[string]interface{} {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_type_id"] = r.DiskTypeId
	res["disk_size"] = toGigabytes(r.DiskSize)

	return []map[string]interface{}{res}
}

func expandSQLServerHost(config map[string]interface{}) (*sqlserver.HostSpec, error) {
	hostSpec := &sqlserver.HostSpec{}
	if v, ok := config["zone"]; ok {
		hostSpec.ZoneId = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		hostSpec.SubnetId = v.(string)
	}

	if v, ok := config["assign_public_ip"]; ok {
		hostSpec.AssignPublicIp = v.(bool)
	}

	return hostSpec, nil
}

func expandSQLServerHosts(d *schema.ResourceData) ([]*sqlserver.HostSpec, error) {
	var hostsSpec []*sqlserver.HostSpec

	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host, err := expandSQLServerHost(config)
		if err != nil {
			return nil, err
		}
		hostsSpec = append(hostsSpec, host)
	}

	return hostsSpec, nil
}

type sqlserverHostKey struct {
	zoneID         string
	subnetID       string
	AssignPublicIP bool
}

func flattenSQLServerHosts(d *schema.ResourceData, hosts []*sqlserver.Host) ([]interface{}, error) {

	hostSpec, err := expandSQLServerHosts(d)

	if err != nil {
		return nil, err
	}

	sortKeys := map[sqlserverHostKey][]int{}

	for i, host := range hostSpec {
		key := sqlserverHostKey{
			zoneID:         host.ZoneId,
			subnetID:       host.SubnetId,
			AssignPublicIP: host.AssignPublicIp,
		}
		if list, ok := sortKeys[key]; ok {
			list = append(list, i)
			sortKeys[key] = list
		} else {
			sortKeys[key] = []int{i}
		}
	}

	result := []interface{}{}

	fqdnOrder := map[string]int{}

	for _, host := range hosts {
		key := sqlserverHostKey{
			zoneID:         host.ZoneId,
			subnetID:       host.SubnetId,
			AssignPublicIP: host.AssignPublicIp,
		}

		if list, ok := sortKeys[key]; ok && len(list) > 0 {
			fqdnOrder[host.Name] = list[0]
			list = list[1:]
			sortKeys[key] = list

		}

		hostFlatten := flattenSQLServerHost(host)

		result = append(result, hostFlatten)
	}

	sort.Slice(result, func(i int, j int) bool {
		return lessInterfaceList(result, "fqdn", i, j, fqdnOrder)
	})

	return result, nil
}

func flattenSQLServerHost(host *sqlserver.Host) map[string]interface{} {
	result := map[string]interface{}{}
	result["fqdn"] = host.Name
	result["zone"] = host.ZoneId
	result["subnet_id"] = host.SubnetId
	result["assign_public_ip"] = host.AssignPublicIp

	return result
}

var rolesSQLServer = map[string]sqlserver.Permission_Role{
	"OWNER":          sqlserver.Permission_DB_OWNER,
	"SECURITYADMIN":  sqlserver.Permission_DB_SECURITYADMIN,
	"ACCESSADMIN":    sqlserver.Permission_DB_ACCESSADMIN,
	"BACKUPOPERATOR": sqlserver.Permission_DB_BACKUPOPERATOR,
	"DDLADMIN":       sqlserver.Permission_DB_DDLADMIN,
	"DATAWRITER":     sqlserver.Permission_DB_DATAWRITER,
	"DATAREADER":     sqlserver.Permission_DB_DATAREADER,
	"DENYDATAWRITER": sqlserver.Permission_DB_DENYDATAWRITER,
	"DENYDATAREADER": sqlserver.Permission_DB_DENYDATAREADER,
}

var rolesRevertedSQLServer = map[sqlserver.Permission_Role]string{
	sqlserver.Permission_DB_OWNER:          "OWNER",
	sqlserver.Permission_DB_SECURITYADMIN:  "SECURITYADMIN",
	sqlserver.Permission_DB_ACCESSADMIN:    "ACCESSADMIN",
	sqlserver.Permission_DB_BACKUPOPERATOR: "BACKUPOPERATOR",
	sqlserver.Permission_DB_DDLADMIN:       "DDLADMIN",
	sqlserver.Permission_DB_DATAWRITER:     "DATAWRITER",
	sqlserver.Permission_DB_DATAREADER:     "DATAREADER",
	sqlserver.Permission_DB_DENYDATAWRITER: "DENYDATAWRITER",
	sqlserver.Permission_DB_DENYDATAREADER: "DENYDATAREADER",
}

func expandSQLServerUserPasswords(d *schema.ResourceData) map[string]string {
	result := map[string]string{}
	users := d.Get("user").([]interface{})

	for _, u := range users {
		m := u.(map[string]interface{})

		result[m["name"].(string)] = m["password"].(string)
	}

	return result
}

func expandSQLServerUserSpecs(d *schema.ResourceData) ([]*sqlserver.UserSpec, error) {
	result := []*sqlserver.UserSpec{}
	users := d.Get("user").([]interface{})

	for _, u := range users {
		m := u.(map[string]interface{})

		user, err := expandSQLServerUser(m)
		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	return result, nil
}

func expandSQLServerUser(u map[string]interface{}) (*sqlserver.UserSpec, error) {
	user := &sqlserver.UserSpec{}

	user.Name = u["name"].(string)
	user.Password = u["password"].(string)

	if v, ok := u["permission"]; ok {
		permissions, err := expandSQLServerUserPermissions(v.(*schema.Set))
		if err != nil {
			return nil, err
		}
		user.Permissions = permissions
	}

	return user, nil
}

func expandSQLServerUserPermissions(permissions *schema.Set) ([]*sqlserver.Permission, error) {
	result := []*sqlserver.Permission{}

	for _, p := range permissions.List() {

		m := p.(map[string]interface{})
		permission := &sqlserver.Permission{}

		if v, ok := m["database_name"]; ok {
			permission.DatabaseName = v.(string)
		}
		if v, ok := m["roles"]; ok && v != nil {
			roles := v.(*schema.Set)
			for _, role := range roles.List() {
				roleSQLServer, ok := rolesSQLServer[role.(string)]
				if ok {
					permission.Roles = append(permission.Roles, roleSQLServer)
				}
			}
		}
		result = append(result, permission)
	}
	return result, nil
}

func flattenSQLServerUsers(users []*sqlserver.User, passwords map[string]string) ([]interface{}, error) {
	result := []interface{}{}

	for _, user := range users {
		userFlatten, err := flattenSQLServerUser(user, passwords[user.Name])
		if err != nil {
			return nil, err
		}
		result = append(result, userFlatten)
	}

	return result, nil
}

func flattenSQLServerUser(user *sqlserver.User, password string) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	result["name"] = user.Name

	if password != "" {
		result["password"] = password
	}

	permissions := []interface{}{}
	for _, permssion := range user.Permissions {
		permissionResult := map[string]interface{}{"database_name": permssion.DatabaseName}
		roles := []interface{}{}
		for _, role := range permssion.Roles {
			roleString, ok := rolesRevertedSQLServer[role]
			if ok {
				roles = append(roles, roleString)
			} else {
				return nil, fmt.Errorf("Unknown SQLServer Permission Role %v", role)
			}
		}

		permissionResult["roles"] = roles

		permissions = append(permissions, permissionResult)
	}

	result["permission"] = permissions

	return result, nil
}

func usersDiffSQLServer(ctx context.Context, config *Config, d *schema.ResourceData) ([]*sqlserver.UserSpec, []*sqlserver.UserSpec, []string, error) {
	oldUsersData, newUsersData := d.GetChange("user")

	passwords := map[string]string{}
	for _, u := range oldUsersData.([]interface{}) {
		m := u.(map[string]interface{})

		passwords[m["name"].(string)] = m["password"].(string)
	}

	newUsers := map[string]*sqlserver.UserSpec{}
	for _, u := range newUsersData.([]interface{}) {
		m := u.(map[string]interface{})

		user, err := expandSQLServerUser(m)
		if err != nil {
			return nil, nil, nil, err
		}
		newUsers[user.Name] = user
	}

	users, err := listSQLServerUsers(ctx, config, d.Id())
	if err != nil {
		return nil, nil, nil, err
	}

	changedUsersSpecs := []*sqlserver.UserSpec{}
	newUsersSpecs := []*sqlserver.UserSpec{}
	dropUserNames := []string{}

	existsUserNames := map[string]struct{}{}

	for _, user := range users {
		newUser, ok := newUsers[user.Name]
		if !ok {
			dropUserNames = append(dropUserNames, user.Name)
			continue
		}
		existsUserNames[user.Name] = struct{}{}
		if password, ok := passwords[user.Name]; !ok || password != newUser.Password {
			changedUsersSpecs = append(changedUsersSpecs, newUser)
			continue
		}

		if !sqlserverUserEqual(user, newUser) {
			changedUsersSpecs = append(changedUsersSpecs, newUser)
		}
	}

	for name, user := range newUsers {
		if _, ok := existsUserNames[name]; !ok {
			newUsersSpecs = append(newUsersSpecs, user)
		}
	}

	return newUsersSpecs, changedUsersSpecs, dropUserNames, nil
}

func sqlserverUserEqual(user *sqlserver.User, userSpec *sqlserver.UserSpec) bool {
	if user.Name != userSpec.Name {
		return false
	}

	if !sqlserverUserPermissionsEqual(user.Permissions, userSpec.Permissions) {
		return false
	}

	return true
}

func sqlserverUserPermissionsEqual(a []*sqlserver.Permission, b []*sqlserver.Permission) bool {

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	keysA := make([]string, 0, len(a))
	keysB := make([]string, 0, len(b))

	for _, permission := range a {
		keysA = append(keysA, sqlserverUserPermissionHashString(permission))
	}

	for _, permission := range b {
		keysB = append(keysB, sqlserverUserPermissionHashString(permission))
	}

	sort.Slice(keysA, func(i, j int) bool {
		return keysA[i] < keysA[j]
	})
	sort.Slice(keysB, func(i, j int) bool {
		return keysB[i] < keysB[j]
	})

	for i := 0; i < len(keysA); i++ {
		if keysA[i] != keysB[i] {
			return false
		}
	}

	return true
}

func sqlserverUserPermissionHashString(permission *sqlserver.Permission) string {
	if permission == nil {
		return ""
	}

	result := permission.DatabaseName

	if len(permission.Roles) == 0 {
		return result
	}

	sort.Slice(permission.Roles, func(i, j int) bool {
		return permission.Roles[i] < permission.Roles[j]
	})

	for _, role := range permission.Roles {
		result += "|" + strconv.Itoa(int(role))
	}

	return result
}

func sqlserverUserPermissionHash(v interface{}) int {
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

func expandSQLServerDatabaseSpecs(d *schema.ResourceData) []*sqlserver.DatabaseSpec {
	result := []*sqlserver.DatabaseSpec{}
	dbs := d.Get("database").([]interface{})

	for _, u := range dbs {
		m := u.(map[string]interface{})

		db := expandSQLServerDatabase(m)
		result = append(result, db)
	}

	return result
}

func expandSQLServerDatabase(u map[string]interface{}) *sqlserver.DatabaseSpec {
	db := &sqlserver.DatabaseSpec{}

	db.Name = u["name"].(string)

	return db
}

func flattenSQLServerDatabases(dbs []*sqlserver.Database) []interface{} {
	result := []interface{}{}

	for _, db := range dbs {
		hostFlatten := flattenSQLServerDatabase(db)

		result = append(result, hostFlatten)
	}

	return result
}

func flattenSQLServerDatabase(db *sqlserver.Database) map[string]interface{} {
	result := map[string]interface{}{}
	result["name"] = db.Name

	return result
}

func databaseDiffSQLServer(ctx context.Context, config *Config, d *schema.ResourceData) ([]*sqlserver.DatabaseSpec, []string, error) {
	databaseData := d.Get("database")

	databasesSpec := map[string]*sqlserver.DatabaseSpec{}
	for _, u := range databaseData.([]interface{}) {
		m := u.(map[string]interface{})

		db := expandSQLServerDatabase(m)
		databasesSpec[db.Name] = db
	}

	dbs, err := listSQLServerDatabases(ctx, config, d.Id())
	if err != nil {
		return nil, nil, err
	}

	newDatabaseSpecs := []*sqlserver.DatabaseSpec{}
	dropDatabaseNames := []string{}

	existsDatabaseNames := map[string]struct{}{}

	for _, db := range dbs {
		_, ok := databasesSpec[db.Name]
		if !ok {
			dropDatabaseNames = append(dropDatabaseNames, db.Name)
			continue
		}
		existsDatabaseNames[db.Name] = struct{}{}

	}

	for name, db := range databasesSpec {
		if _, ok := existsDatabaseNames[name]; !ok {
			newDatabaseSpecs = append(newDatabaseSpecs, db)
		}
	}

	return newDatabaseSpecs, dropDatabaseNames, nil
}

func flattenSQLServerSettings(c *sqlserver.ClusterConfig) (map[string]string, error) {

	if cf, ok := c.SqlserverConfig.(*sqlserver.ClusterConfig_SqlserverConfig_2016Sp2Std); ok {

		settings, err := flattenResourceGenerateMapS(cf.SqlserverConfig_2016Sp2Std.UserConfig, false, mdbSQLServerSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		return settings, nil
	}
	if cf, ok := c.SqlserverConfig.(*sqlserver.ClusterConfig_SqlserverConfig_2016Sp2Ent); ok {
		settings, err := flattenResourceGenerateMapS(cf.SqlserverConfig_2016Sp2Ent.UserConfig, false, mdbSQLServerSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		return settings, nil
	}

	return nil, nil
}

func expandSQLServerConfigSpecSettings(d *schema.ResourceData, configSpec *sqlserver.ConfigSpec) (updateFieldConfigName string, fields []string, err error) {

	version := configSpec.Version

	path := "sqlserver_config"

	if _, ok := d.GetOkExists(path); !ok {
		return "", nil, nil
	}

	var sqlserverConfig interface{}
	if version == "2016sp2std" {
		cfg := &sqlserver.ConfigSpec_SqlserverConfig_2016Sp2Std{
			SqlserverConfig_2016Sp2Std: &config.SQLServerConfig2016Sp2Std{},
		}
		sqlserverConfig = cfg.SqlserverConfig_2016Sp2Std
		configSpec.SqlserverConfig = cfg
		updateFieldConfigName = "sqlserver_config_2016sp2std"
	} else if version == "2016sp2ent" {
		cfg := &sqlserver.ConfigSpec_SqlserverConfig_2016Sp2Ent{
			SqlserverConfig_2016Sp2Ent: &config.SQLServerConfig2016Sp2Ent{},
		}
		sqlserverConfig = cfg.SqlserverConfig_2016Sp2Ent
		configSpec.SqlserverConfig = cfg
		updateFieldConfigName = "sqlserver_config_2016sp2ent"

	} else {
		return "", nil, nil
	}

	fields, err = expandResourceGenerateNonSkippedFields(mdbSQLServerSettingsFieldsInfo, d, sqlserverConfig, path+".", true)

	if err != nil {
		return "", nil, err
	}

	return updateFieldConfigName, fields, nil
}

var mdbSQLServerSettingsFieldsInfo = newObjectFieldsInfo().
	addType(config.SQLServerConfig2016Sp2Std{}).
	addType(config.SQLServerConfig2016Sp2Ent{})
