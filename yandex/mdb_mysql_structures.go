package yandex

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
)

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

func expandMysqlUser(u map[string]interface{}) *mysql.UserSpec {
	user := &mysql.UserSpec{}

	if v, ok := u["name"]; ok {
		user.Name = v.(string)
	}

	if v, ok := u["password"]; ok {
		user.Password = v.(string)
	}

	if v, ok := u["permission"]; ok {
		user.Permissions = expandMysqlUserPermissions(v.(*schema.Set))
	}

	return user
}

func expandMysqlUserSpecs(d *schema.ResourceData) ([]*mysql.UserSpec, error) {
	result := []*mysql.UserSpec{}
	users := d.Get("user").(*schema.Set)

	for _, u := range users.List() {
		m := u.(map[string]interface{})

		result = append(result, expandMysqlUser(m))
	}

	return result, nil
}

func expandMysqlUserPermissions(ps *schema.Set) []*mysql.Permission {
	result := []*mysql.Permission{}

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		permission := &mysql.Permission{}
		if v, ok := m["database_name"]; ok {
			permission.DatabaseName = v.(string)
		}
		result = append(result, permission)
	}
	return result
}

func expandMysqlHosts(d *schema.ResourceData) ([]*mysql.HostSpec, error) {
	var result []*mysql.HostSpec
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

func expandMysqlHost(config map[string]interface{}) (*mysql.HostSpec, error) {
	host := &mysql.HostSpec{}
	if v, ok := config["zone"]; ok {
		host.ZoneId = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		host.SubnetId = v.(string)
	}

	if v, ok := config["assign_public_ip"]; ok {
		host.AssignPublicIp = v.(bool)
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

func mysqlUserHash(v interface{}) int {
	buf := bytes.Buffer{}

	m := v.(map[string]interface{})
	if n, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", n.(string)))
	}
	if p, ok := m["password"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", p.(string)))
	}
	if ps, ok := m["permission"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", ps.(*schema.Set).List()))
	}

	return hashcode.String(buf.String())
}

func mysqlDatabaseHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func mysqlUserPermissionHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["database_name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
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

	if v, ok := d.GetOk("config.0.backup_window_start.0.hours"); ok {
		out.Hours = int32(v.(int))
	}

	if v, ok := d.GetOk("config.0.backup_window_start.0.minutes"); ok {
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

func sortMysqlHosts(hosts []*mysql.Host, specs []*mysql.HostSpec) {
	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if h.ZoneId == hosts[j].ZoneId {
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

func flattenMysqlUsers(us []*mysql.User, passwords map[string]string) (*schema.Set, error) {
	out := schema.NewSet(mysqlUserHash, nil)

	for _, u := range us {
		ou, err := flattenMysqlUser(u)
		if err != nil {
			return nil, err
		}

		if v, ok := passwords[u.Name]; ok {
			ou["password"] = v
		}

		out.Add(ou)
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

	return m, nil
}

func flattenMysqlUserPermissions(ps []*mysql.Permission) (*schema.Set, error) {
	out := schema.NewSet(mysqlUserPermissionHash, nil)

	for _, p := range ps {
		op := map[string]interface{}{
			"database_name": p.DatabaseName,
		}

		out.Add(op)
	}

	return out, nil
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
