package yandex

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
)

type mongodbConfig struct {
	version                     string
	featureCompatibilityVersion string
	backupWindowStart           *map[string]interface{}
	access                      *mongodb.Access
}

func extractMongoDBConfig(cc *mongodb.ClusterConfig) mongodbConfig {
	t := cc.BackupWindowStart

	r := map[string]interface{}{}

	r["hours"] = int(t.Hours)
	r["minutes"] = int(t.Minutes)

	res := mongodbConfig{}
	res.version = cc.Version
	res.featureCompatibilityVersion = cc.FeatureCompatibilityVersion
	res.backupWindowStart = &r
	res.access = cc.Access
	return res
}

func flattenMongoDBResources(m *mongodb.Resources) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	res["resource_preset_id"] = m.ResourcePresetId
	res["disk_size"] = toGigabytes(m.DiskSize)
	res["disk_type_id"] = m.DiskTypeId

	return []map[string]interface{}{res}, nil
}

func flattenMongoDBHosts(hs []*mongodb.Host) ([]map[string]interface{}, error) {
	res := []map[string]interface{}{}

	for _, h := range hs {
		m := map[string]interface{}{}
		m["zone_id"] = h.ZoneId
		m["subnet_id"] = h.SubnetId
		m["name"] = h.Name
		m["role"] = h.Role.String()
		m["health"] = h.Health.String()
		m["assign_public_ip"] = h.AssignPublicIp
		m["shard_name"] = h.ShardName
		m["type"] = h.Type.String()
		res = append(res, m)
	}

	return res, nil
}

func expandMongoDBHosts(d *schema.ResourceData) ([]*mongodb.HostSpec, error) {
	var result []*mongodb.HostSpec
	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host := expandMongoDBHost(config)
		result = append(result, host)
	}

	return result, nil
}

func expandMongoDBHost(config map[string]interface{}) *mongodb.HostSpec {
	host := &mongodb.HostSpec{}
	if v, ok := config["type"]; ok {
		host.Type = mongodb.Host_Type(mongodb.Host_Type_value[v.(string)])
	}

	if v, ok := config["zone_id"]; ok {
		host.ZoneId = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		host.SubnetId = v.(string)
	}

	if v, ok := config["shard_name"]; ok {
		host.ShardName = v.(string)
	}

	if v, ok := config["assign_public_ip"]; ok {
		host.AssignPublicIp = v.(bool)
	}
	return host
}

func parseMongoDBEnv(e string) (mongodb.Cluster_Environment, error) {
	v, ok := mongodb.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(mongodb.Cluster_Environment_value)), e)
	}
	return mongodb.Cluster_Environment(v), nil
}

func mongodbUserPermissionHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["database_name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func mongodbUserHash(v interface{}) int {
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

func mongodbDatabaseHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func mongodbUsersPasswords(users []*mongodb.UserSpec) map[string]string {
	result := map[string]string{}
	for _, u := range users {
		result[u.Name] = u.Password
	}
	return result
}

func flattenMongoDBUsers(users []*mongodb.User, passwords map[string]string) *schema.Set {
	result := schema.NewSet(mongodbUserHash, nil)

	for _, user := range users {
		u := map[string]interface{}{}
		u["name"] = user.Name

		perms := schema.NewSet(mongodbUserPermissionHash, nil)
		for _, perm := range user.Permissions {
			p := map[string]interface{}{}
			p["database_name"] = perm.DatabaseName
			p["roles"] = perm.Roles
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

func flattenMongoDBDatabases(dbs []*mongodb.Database) *schema.Set {
	result := schema.NewSet(mongodbDatabaseHash, nil)

	for _, d := range dbs {
		m := make(map[string]interface{})
		m["name"] = d.Name
		result.Add(m)
	}
	return result
}

func expandMongoDBUser(u map[string]interface{}) *mongodb.UserSpec {
	user := &mongodb.UserSpec{}

	if v, ok := u["name"]; ok {
		user.Name = v.(string)
	}

	if v, ok := u["password"]; ok {
		user.Password = v.(string)
	}

	if v, ok := u["permission"]; ok {
		user.Permissions = expandMongoDBUserPermissions(v.(*schema.Set))
	}

	return user
}

func expandMongoDBUserSpecs(d *schema.ResourceData) ([]*mongodb.UserSpec, error) {
	result := []*mongodb.UserSpec{}
	users := d.Get("user").(*schema.Set)

	for _, u := range users.List() {
		m := u.(map[string]interface{})

		result = append(result, expandMongoDBUser(m))
	}

	return result, nil
}

func expandMongoDBUserPermissions(ps *schema.Set) []*mongodb.Permission {
	result := []*mongodb.Permission{}

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		permission := &mongodb.Permission{}
		if v, ok := m["database_name"]; ok {
			permission.DatabaseName = v.(string)
		}

		if v, ok := m["roles"]; ok {
			roles := make([]string, len(v.([]interface{})))
			for n, item := range v.([]interface{}) {
				roles[n] = item.(string)
			}

			permission.Roles = roles
		}
		result = append(result, permission)
	}
	return result
}

func expandMongoDBDatabases(d *schema.ResourceData) ([]*mongodb.DatabaseSpec, error) {
	var result []*mongodb.DatabaseSpec
	dbs := d.Get("database").(*schema.Set).List()

	for _, d := range dbs {
		m := d.(map[string]interface{})
		db := &mongodb.DatabaseSpec{}

		if v, ok := m["name"]; ok {
			db.Name = v.(string)
		}

		result = append(result, db)
	}
	return result, nil
}

func expandMongoDBResources(d *schema.ResourceData) *mongodb.Resources {
	res := mongodb.Resources{
		DiskSize:         toBytes(d.Get("resources.0.disk_size").(int)),
		DiskTypeId:       d.Get("resources.0.disk_type_id").(string),
		ResourcePresetId: d.Get("resources.0.resource_preset_id").(string),
	}

	return &res
}

func expandMongoDBBackupWindowStart(d *schema.ResourceData) *timeofday.TimeOfDay {
	res := timeofday.TimeOfDay{
		Hours:   int32(d.Get("cluster_config.0.backup_window_start.0.hours").(int)),
		Minutes: int32(d.Get("cluster_config.0.backup_window_start.0.minutes").(int)),
	}

	return &res
}

//the following expansion works only because sharded mongodb is not supported

func expandMongoDBSpec4_2(d *schema.ResourceData) *mongodb.ConfigSpec_MongodbSpec_4_2 {
	return &mongodb.ConfigSpec_MongodbSpec_4_2{
		MongodbSpec_4_2: &mongodb.MongodbSpec4_2{
			Mongod: &mongodb.MongodbSpec4_2_Mongod{
				Resources: expandMongoDBResources(d),
			},
			Mongos:   &mongodb.MongodbSpec4_2_Mongos{},
			Mongocfg: &mongodb.MongodbSpec4_2_MongoCfg{},
		},
	}
}

func expandMongoDBSpec4_0(d *schema.ResourceData) *mongodb.ConfigSpec_MongodbSpec_4_0 {
	return &mongodb.ConfigSpec_MongodbSpec_4_0{
		MongodbSpec_4_0: &mongodb.MongodbSpec4_0{
			Mongod: &mongodb.MongodbSpec4_0_Mongod{
				Resources: expandMongoDBResources(d),
			},
			Mongos:   &mongodb.MongodbSpec4_0_Mongos{},
			Mongocfg: &mongodb.MongodbSpec4_0_MongoCfg{},
		},
	}
}

func expandMongoDBSpec3_6(d *schema.ResourceData) *mongodb.ConfigSpec_MongodbSpec_3_6 {
	return &mongodb.ConfigSpec_MongodbSpec_3_6{
		MongodbSpec_3_6: &mongodb.MongodbSpec3_6{
			Mongod: &mongodb.MongodbSpec3_6_Mongod{
				Resources: expandMongoDBResources(d),
			},
			Mongos:   &mongodb.MongodbSpec3_6_Mongos{},
			Mongocfg: &mongodb.MongodbSpec3_6_MongoCfg{},
		},
	}
}

func mongodbDatabasesDiff(currDBs []*mongodb.Database, targetDBs []*mongodb.DatabaseSpec) ([]string, []string) {
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
