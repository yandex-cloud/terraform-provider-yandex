package yandex

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	timeofday "google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

// Sorts list of hosts in accordance with the order in config.
// We need to keep the original order so there's no diff appears on each apply.
// Removes implicit ZooKeeper hosts from the `hosts` slice.
func sortClickHouseHosts(hosts []*clickhouse.Host, specs []*clickhouse.HostSpec) []*clickhouse.Host {
	implicitZk := true
	for _, h := range specs {
		if h.Type == clickhouse.Host_ZOOKEEPER {
			implicitZk = false
			break
		}
	}

	if implicitZk {
		n := 0
		for _, h := range hosts {
			// Filter out implicit ZooKeeper hosts.
			if h.Type == clickhouse.Host_CLICKHOUSE {
				hosts[n] = h
				n++
			}
		}
		hosts = hosts[:n]
	}

	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if h.ZoneId == hosts[j].ZoneId && (h.ShardName == "" || h.ShardName == hosts[j].ShardName) && h.Type == hosts[j].Type {
				hosts[i], hosts[j] = hosts[j], hosts[i]
				break
			}
		}
	}
	return hosts
}

func clickHouseUserPermissionHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["database_name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func clickHouseUserHash(v interface{}) int {
	var buf bytes.Buffer

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

func clickHouseDatabaseHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

// Takes the current list of dbs and the desirable list of dbs.
// Returns the slice of dbs to delete and the slice of dbs to add.
func clickHouseDatabasesDiff(currDBs []*clickhouse.Database, targetDBs []*clickhouse.DatabaseSpec) ([]string, []string) {
	m := map[string]bool{}
	toAdd := []string{}
	toDelete := map[string]bool{}
	for _, db := range currDBs {
		toDelete[db.Name] = true
		m[db.Name] = true
	}

	for _, db := range targetDBs {
		delete(toDelete, db.Name)
		if _, ok := m[db.Name]; !ok {
			toAdd = append(toAdd, db.Name)
		}
	}

	toDel := []string{}
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

// Takes the current list of users and the desirable list of users.
// Returns the slice of usernames to delete and the slice of users to add.
func clickHouseUsersDiff(currUsers []*clickhouse.User, targetUsers []*clickhouse.UserSpec) ([]string, []*clickhouse.UserSpec) {
	m := map[string]bool{}
	toDelete := map[string]bool{}
	toAdd := []*clickhouse.UserSpec{}

	for _, u := range currUsers {
		toDelete[u.Name] = true
		m[u.Name] = true
	}

	for _, u := range targetUsers {
		delete(toDelete, u.Name)
		if _, ok := m[u.Name]; !ok {
			toAdd = append(toAdd, u)
		}
	}

	toDel := []string{}
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

// Takes the old set of user specs and the new set of user specs.
// Returns the slice of user specs which have changed.
func clickHouseChangedUsers(oldSpecs *schema.Set, newSpecs *schema.Set) []*clickhouse.UserSpec {
	result := []*clickhouse.UserSpec{}
	m := map[string]*clickhouse.UserSpec{}
	for _, spec := range oldSpecs.List() {
		user := expandClickHouseUser(spec.(map[string]interface{}))
		m[user.Name] = user
	}
	for _, spec := range newSpecs.List() {
		user := expandClickHouseUser(spec.(map[string]interface{}))
		if u, ok := m[user.Name]; ok {
			if user.Password != u.Password || fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) {
				result = append(result, user)
			}
		}
	}
	return result
}

// Takes the current list of hosts and the desirable list of hosts.
// Returns the map of hostnames to delete grouped by shard,
// and the map of hosts to add grouped by shard as well.
// All the ZOOKEEPER hosts will reside under the key "zk".
func clickHouseHostsDiff(currHosts []*clickhouse.Host, targetHosts []*clickhouse.HostSpec) (map[string][]string, map[string][]*clickhouse.HostSpec) {
	m := map[string][]*clickhouse.HostSpec{}

	for _, h := range targetHosts {
		shardName := "shard1"
		if h.ShardName != "" {
			shardName = h.ShardName
		}
		if h.Type == clickhouse.Host_ZOOKEEPER {
			shardName = "zk"
		}
		key := h.Type.String() + h.ZoneId + shardName
		m[key] = append(m[key], h)
	}

	toDelete := map[string][]string{}
	for _, h := range currHosts {
		shardName := h.ShardName
		if h.Type == clickhouse.Host_ZOOKEEPER {
			shardName = "zk"
		}
		key := h.Type.String() + h.ZoneId + shardName
		hs, ok := m[key]
		if !ok {
			toDelete[shardName] = append(toDelete[h.ShardName], h.Name)
		}
		if len(hs) > 1 {
			m[key] = hs[1:]
		} else {
			delete(m, key)
		}
	}

	toAdd := map[string][]*clickhouse.HostSpec{}
	for _, hs := range m {
		for _, h := range hs {
			if h.Type == clickhouse.Host_ZOOKEEPER {
				toAdd["zk"] = append(toAdd["zk"], h)
			} else {
				toAdd[h.ShardName] = append(toAdd[h.ShardName], h)
			}
		}
	}

	return toDelete, toAdd
}

func parseClickHouseEnv(e string) (clickhouse.Cluster_Environment, error) {
	v, ok := clickhouse.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(clickhouse.Cluster_Environment_value)), e)
	}
	return clickhouse.Cluster_Environment(v), nil
}

func parseClickHouseHostType(t string) (clickhouse.Host_Type, error) {
	v, ok := clickhouse.Host_Type_value[t]
	if !ok {
		return 0, fmt.Errorf("value for 'host.type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(clickhouse.Host_Type_value)), t)
	}
	return clickhouse.Host_Type(v), nil
}

func expandClickHouseHosts(d *schema.ResourceData) ([]*clickhouse.HostSpec, error) {
	var result []*clickhouse.HostSpec
	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host, err := expandClickHouseHost(config)
		if err != nil {
			return nil, err
		}
		result = append(result, host)
	}

	return result, nil
}

func expandClickHouseHost(config map[string]interface{}) (*clickhouse.HostSpec, error) {
	host := &clickhouse.HostSpec{}
	if v, ok := config["zone"]; ok {
		host.ZoneId = v.(string)
	}

	if v, ok := config["type"]; ok {
		t, err := parseClickHouseHostType(v.(string))
		if err != nil {
			return nil, err
		}
		host.Type = t
	}

	if v, ok := config["subnet_id"]; ok {
		host.SubnetId = v.(string)
	}

	if v, ok := config["shard_name"]; ok {
		host.ShardName = v.(string)
		if host.Type == clickhouse.Host_ZOOKEEPER && host.ShardName != "" {
			return nil, fmt.Errorf("ZooKeeper hosts cannot have a 'shard_name'")
		}
	}

	if v, ok := config["assign_public_ip"]; ok {
		host.AssignPublicIp = v.(bool)
	}

	return host, nil
}

func flattenClickHouseDatabases(dbs []*clickhouse.Database) *schema.Set {
	result := schema.NewSet(clickHouseDatabaseHash, nil)

	for _, d := range dbs {
		m := make(map[string]interface{})
		m["name"] = d.Name
		result.Add(m)
	}
	return result
}

func flattenClickHouseResources(r *clickhouse.Resources) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_type_id"] = r.DiskTypeId
	res["disk_size"] = toGigabytes(r.DiskSize)

	return []map[string]interface{}{res}, nil
}

func expandClickHouseResources(d *schema.ResourceData, rootKey string) *clickhouse.Resources {
	resources := &clickhouse.Resources{}

	if v, ok := d.GetOk(rootKey + ".resource_preset_id"); ok {
		resources.ResourcePresetId = v.(string)
	}
	if v, ok := d.GetOk(rootKey + ".disk_size"); ok {
		resources.DiskSize = toBytes(v.(int))
	}
	if v, ok := d.GetOk(rootKey + ".disk_type_id"); ok {
		resources.DiskTypeId = v.(string)
	}
	return resources
}

func expandClickHouseZookeeperSpec(d *schema.ResourceData) *clickhouse.ConfigSpec_Zookeeper {
	result := &clickhouse.ConfigSpec_Zookeeper{}
	result.Resources = expandClickHouseResources(d, "zookeeper.0.resources.0")
	return result
}

func expandClickHouseSpec(d *schema.ResourceData) *clickhouse.ConfigSpec_Clickhouse {
	result := &clickhouse.ConfigSpec_Clickhouse{}
	result.Resources = expandClickHouseResources(d, "clickhouse.0.resources.0")

	return result
}

func flattenClickHouseBackupWindowStart(t *timeofday.TimeOfDay) []map[string]interface{} {
	res := map[string]interface{}{}

	res["hours"] = int(t.Hours)
	res["minutes"] = int(t.Minutes)

	return []map[string]interface{}{res}
}

func expandClickHouseBackupWindowStart(d *schema.ResourceData) *timeofday.TimeOfDay {
	result := &timeofday.TimeOfDay{}

	if v, ok := d.GetOk("backup_window_start.0.hours"); ok {
		result.Hours = int32(v.(int))
	}
	if v, ok := d.GetOk("backup_window_start.0.minutes"); ok {
		result.Minutes = int32(v.(int))
	}
	return result
}

func flattenClickHouseAccess(a *clickhouse.Access) []map[string]interface{} {
	res := map[string]interface{}{}

	res["web_sql"] = a.WebSql
	res["data_lens"] = a.DataLens
	res["metrika"] = a.Metrika
	res["serverless"] = a.Serverless

	return []map[string]interface{}{res}
}

func expandClickHouseAccess(d *schema.ResourceData) *clickhouse.Access {
	result := &clickhouse.Access{}

	if v, ok := d.GetOk("access.0.web_sql"); ok {
		result.WebSql = v.(bool)
	}
	if v, ok := d.GetOk("access.0.data_lens"); ok {
		result.DataLens = v.(bool)
	}
	if v, ok := d.GetOk("access.0.metrika"); ok {
		result.Metrika = v.(bool)
	}
	if v, ok := d.GetOk("access.0.serverless"); ok {
		result.Serverless = v.(bool)
	}
	return result
}

func expandClickHouseUserPermissions(ps *schema.Set) []*clickhouse.Permission {
	result := []*clickhouse.Permission{}

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		permission := &clickhouse.Permission{}
		if v, ok := m["database_name"]; ok {
			permission.DatabaseName = v.(string)
		}
		result = append(result, permission)
	}
	return result
}

func flattenClickHouseUsers(users []*clickhouse.User, passwords map[string]string) *schema.Set {
	result := schema.NewSet(clickHouseUserHash, nil)

	for _, user := range users {
		u := map[string]interface{}{}
		u["name"] = user.Name

		perms := schema.NewSet(clickHouseUserPermissionHash, nil)
		for _, perm := range user.Permissions {
			p := map[string]interface{}{}
			p["database_name"] = perm.DatabaseName
			perms.Add(p)
		}
		u["permission"] = perms

		if p, ok := passwords[user.Name]; ok {
			u["password"] = p
		}
		result.Add(u)
	}
	return result
}

func expandClickHouseUser(u map[string]interface{}) *clickhouse.UserSpec {
	user := &clickhouse.UserSpec{}

	if v, ok := u["name"]; ok {
		user.Name = v.(string)
	}

	if v, ok := u["password"]; ok {
		user.Password = v.(string)
	}

	if v, ok := u["permission"]; ok {
		user.Permissions = expandClickHouseUserPermissions(v.(*schema.Set))
	}

	return user
}

func expandClickHouseUserSpecs(d *schema.ResourceData) ([]*clickhouse.UserSpec, error) {
	result := []*clickhouse.UserSpec{}
	users := d.Get("user").(*schema.Set)

	for _, u := range users.List() {
		m := u.(map[string]interface{})

		result = append(result, expandClickHouseUser(m))
	}

	return result, nil
}

func clickHouseUsersPasswords(users []*clickhouse.UserSpec) map[string]string {
	result := map[string]string{}
	for _, u := range users {
		result[u.Name] = u.Password
	}
	return result
}

func expandClickHouseDatabases(d *schema.ResourceData) ([]*clickhouse.DatabaseSpec, error) {
	var result []*clickhouse.DatabaseSpec
	dbs := d.Get("database").(*schema.Set).List()

	for _, d := range dbs {
		m := d.(map[string]interface{})
		db := &clickhouse.DatabaseSpec{}

		if v, ok := m["name"]; ok {
			db.Name = v.(string)
		}

		result = append(result, db)
	}
	return result, nil
}

func flattenClickHouseHosts(hs []*clickhouse.Host) ([]map[string]interface{}, error) {
	res := []map[string]interface{}{}

	for _, h := range hs {
		m := map[string]interface{}{}
		m["type"] = h.GetType().String()
		m["zone"] = h.ZoneId
		m["subnet_id"] = h.SubnetId
		m["shard_name"] = h.ShardName
		m["assign_public_ip"] = h.AssignPublicIp
		m["fqdn"] = h.Name
		res = append(res, m)
	}

	return res, nil
}
