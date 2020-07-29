package yandex

import (
	"bytes"
	"fmt"
	"reflect"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	timeofday "google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

type PostgreSQLHostSpec struct {
	HostSpec        *postgresql.HostSpec
	Fqdn            string
	HasComputedFqdn bool
}

func flattenPGClusterConfig(c *postgresql.ClusterConfig) ([]interface{}, error) {
	poolerConf, err := flattenPGPoolerConfig(c.PoolerConfig)
	if err != nil {
		return nil, err
	}

	resources, err := flattenPGResources(c.Resources)
	if err != nil {
		return nil, err
	}

	backupWindowStart, err := flattenPGBackupWindowStart(c.BackupWindowStart)
	if err != nil {
		return nil, err
	}

	access, err := flattenPGAccess(c.Access)
	if err != nil {
		return nil, err
	}

	out := map[string]interface{}{}
	out["autofailover"] = c.GetAutofailover().GetValue()
	out["version"] = c.Version
	out["pooler_config"] = poolerConf
	out["resources"] = resources
	out["backup_window_start"] = backupWindowStart
	out["access"] = access

	return []interface{}{out}, nil
}

func flattenPGPoolerConfig(c *postgresql.ConnectionPoolerConfig) ([]interface{}, error) {
	if c == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["pool_discard"] = c.GetPoolDiscard().GetValue()
	out["pooling_mode"] = c.GetPoolingMode().String()

	return []interface{}{out}, nil
}

func flattenPGResources(r *postgresql.Resources) ([]interface{}, error) {
	out := map[string]interface{}{}
	out["resource_preset_id"] = r.ResourcePresetId
	out["disk_size"] = toGigabytes(r.DiskSize)
	out["disk_type_id"] = r.DiskTypeId

	return []interface{}{out}, nil
}

func flattenPGBackupWindowStart(t *timeofday.TimeOfDay) ([]interface{}, error) {
	if t == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["hours"] = int(t.Hours)
	out["minutes"] = int(t.Minutes)

	return []interface{}{out}, nil
}

func flattenPGAccess(a *postgresql.Access) ([]interface{}, error) {
	if a == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["data_lens"] = a.DataLens

	return []interface{}{out}, nil
}

func flattenPGUsers(us []*postgresql.User, passwords map[string]string) ([]map[string]interface{}, error) {
	out := make([]map[string]interface{}, 0)

	for _, u := range us {
		ou, err := flattenPGUser(u)
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

func flattenPGUser(u *postgresql.User) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	m["name"] = u.Name
	m["login"] = u.GetLogin().GetValue()

	permissions, err := flattenPGUserPermissions(u.Permissions)
	if err != nil {
		return nil, err
	}
	m["permission"] = permissions

	m["grants"] = u.Grants

	return m, nil
}

func pgUsersPasswords(users []*postgresql.UserSpec) map[string]string {
	out := map[string]string{}
	for _, u := range users {
		out[u.Name] = u.Password
	}
	return out
}

func pgUserPermissionHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["database_name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func flattenPGUserPermissions(ps []*postgresql.Permission) (*schema.Set, error) {
	out := schema.NewSet(pgUserPermissionHash, nil)

	for _, p := range ps {
		op := map[string]interface{}{
			"database_name": p.DatabaseName,
		}

		out.Add(op)
	}

	return out, nil
}

func flattenPGHosts(hs []*postgresql.Host) ([]map[string]interface{}, error) {
	out := []map[string]interface{}{}

	for _, h := range hs {
		m := map[string]interface{}{}
		m["zone"] = h.ZoneId
		m["subnet_id"] = h.SubnetId
		m["assign_public_ip"] = h.AssignPublicIp
		m["fqdn"] = h.Name
		m["role"] = h.Role.String()

		out = append(out, m)
	}

	return out, nil
}

func flattenPGDatabases(dbs []*postgresql.Database) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)

	for _, d := range dbs {
		m := make(map[string]interface{})
		m["name"] = d.Name
		m["owner"] = d.Owner
		m["lc_collate"] = d.LcCollate
		m["lc_type"] = d.LcCtype
		m["extension"] = flattenPGExtensions(d.Extensions)

		out = append(out, m)
	}

	return out
}

func flattenPGExtensions(es []*postgresql.Extension) *schema.Set {
	out := schema.NewSet(pgExtensionHash, nil)

	for _, e := range es {
		m := make(map[string]interface{})
		m["name"] = e.Name
		m["version"] = e.Version

		out.Add(m)
	}

	return out
}

func pgExtensionHash(v interface{}) int {
	var buf bytes.Buffer

	m := v.(map[string]interface{})
	if v, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["version"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}

func expandPGConfigSpec(d *schema.ResourceData) (*postgresql.ConfigSpec, error) {
	cs := &postgresql.ConfigSpec{}

	if v, ok := d.GetOk("config.0.version"); ok {
		cs.Version = v.(string)
	}

	if v, ok := d.GetOkExists("config.0.autofailover"); ok {
		cs.Autofailover = &wrappers.BoolValue{Value: v.(bool)}
	}

	poolerConfig, err := expandPGPoolerConfig(d)
	if err != nil {
		return nil, err
	}
	cs.PoolerConfig = poolerConfig

	resources, err := expandPGResources(d)
	if err != nil {
		return nil, err
	}
	cs.Resources = resources

	cs.BackupWindowStart = expandPGBackupWindowStart(d)
	cs.Access = expandPGAccess(d)

	return cs, nil
}

func expandPGPoolerConfig(d *schema.ResourceData) (*postgresql.ConnectionPoolerConfig, error) {
	pc := &postgresql.ConnectionPoolerConfig{}

	if v, ok := d.GetOk("config.0.pooler_config.0.pooling_mode"); ok {
		pm, err := parsePostgreSQLPoolingMode(v.(string))
		if err != nil {
			return nil, err
		}

		pc.PoolingMode = pm
	}

	if v, ok := d.GetOk("config.0.pooler_config.0.pool_discard"); ok {
		pc.PoolDiscard = &wrappers.BoolValue{Value: v.(bool)}
	}

	return pc, nil
}

func expandPGResources(d *schema.ResourceData) (*postgresql.Resources, error) {
	r := &postgresql.Resources{}

	if v, ok := d.GetOk("config.0.resources.0.resource_preset_id"); ok {
		r.ResourcePresetId = v.(string)
	}

	if v, ok := d.GetOk("config.0.resources.0.disk_size"); ok {
		r.DiskSize = toBytes(v.(int))
	}

	if v, ok := d.GetOk("config.0.resources.0.disk_type_id"); ok {
		r.DiskTypeId = v.(string)
	}

	return r, nil
}

func expandPGUserSpecs(d *schema.ResourceData) ([]*postgresql.UserSpec, error) {
	out := []*postgresql.UserSpec{}
	users := d.Get("user").([]interface{})

	for _, u := range users {
		m := u.(map[string]interface{})

		user, err := expandPGUser(m)
		if err != nil {
			return nil, err
		}

		out = append(out, user)
	}

	return out, nil
}

func expandPGUser(m map[string]interface{}) (*postgresql.UserSpec, error) {
	user := &postgresql.UserSpec{}

	if v, ok := m["name"]; ok {
		user.Name = v.(string)
	}

	if v, ok := m["password"]; ok {
		user.Password = v.(string)
	}

	if v, ok := m["login"]; ok {
		user.Login = &wrappers.BoolValue{Value: v.(bool)}
	}

	if v, ok := m["permission"]; ok {
		permissions, err := expandPGUserPermissions(v.(*schema.Set))
		if err != nil {
			return nil, err
		}
		user.Permissions = permissions
	}

	if v, ok := m["grants"]; ok {
		gs, err := expandPGUserGrants(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		user.Grants = gs
	}

	return user, nil
}

func expandPGUserGrants(gs []interface{}) ([]string, error) {
	out := []string{}

	if gs == nil {
		return out, nil
	}

	for _, v := range gs {
		out = append(out, v.(string))
	}

	return out, nil
}

func expandPGUserPermissions(ps *schema.Set) ([]*postgresql.Permission, error) {
	out := []*postgresql.Permission{}

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		permission := &postgresql.Permission{}

		if v, ok := m["database_name"]; ok {
			permission.DatabaseName = v.(string)
		}

		out = append(out, permission)
	}

	return out, nil
}

func expandPGHosts(d *schema.ResourceData) ([]*PostgreSQLHostSpec, error) {
	out := []*PostgreSQLHostSpec{}
	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		m := v.(map[string]interface{})
		h, err := expandPGHost(m)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}

	return out, nil
}

func expandPGHost(m map[string]interface{}) (*PostgreSQLHostSpec, error) {
	hostSpec := &postgresql.HostSpec{}
	host := &PostgreSQLHostSpec{HostSpec: hostSpec, HasComputedFqdn: false}
	if v, ok := m["zone"]; ok {
		host.HostSpec.ZoneId = v.(string)
	}

	if v, ok := m["subnet_id"]; ok {
		host.HostSpec.SubnetId = v.(string)
	}

	if v, ok := m["assign_public_ip"]; ok {
		host.HostSpec.AssignPublicIp = v.(bool)
	}
	if v, ok := m["fqdn"]; ok && v.(string) != "" {
		host.HasComputedFqdn = true
		host.Fqdn = v.(string)
	}

	return host, nil
}

func sortPGHosts(hosts []*postgresql.Host, specs []*PostgreSQLHostSpec) {
	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if h.HostSpec.ZoneId == hosts[j].ZoneId {
				hosts[i], hosts[j] = hosts[j], hosts[i]
				break
			}
		}
	}
}

func expandPGDatabaseSpecs(d *schema.ResourceData) ([]*postgresql.DatabaseSpec, error) {
	out := []*postgresql.DatabaseSpec{}
	dbs := d.Get("database").([]interface{})

	for _, d := range dbs {
		m := d.(map[string]interface{})
		database, err := expandPGDatabase(m)
		if err != nil {
			return nil, err
		}

		out = append(out, database)
	}

	return out, nil
}

func expandPGDatabase(m map[string]interface{}) (*postgresql.DatabaseSpec, error) {
	out := &postgresql.DatabaseSpec{}

	if v, ok := m["name"]; ok {
		out.Name = v.(string)
	}

	if v, ok := m["owner"]; ok {
		out.Owner = v.(string)
	}

	if v, ok := m["lc_collate"]; ok {
		out.LcCollate = v.(string)
	}

	if v, ok := m["lc_type"]; ok {
		out.LcCtype = v.(string)
	}

	if v, ok := m["extension"]; ok {
		es := v.(*schema.Set).List()
		extensions, err := expandPGExtensions(es)
		if err != nil {
			return nil, err
		}

		out.Extensions = extensions
	}

	return out, nil
}

func expandPGExtensions(es []interface{}) ([]*postgresql.Extension, error) {
	out := []*postgresql.Extension{}

	for _, e := range es {
		m := e.(map[string]interface{})
		extension := &postgresql.Extension{}

		if v, ok := m["name"]; ok {
			extension.Name = v.(string)
		}

		if v, ok := m["version"]; ok {
			extension.Version = v.(string)
		}

		out = append(out, extension)
	}

	return out, nil
}

func expandPGBackupWindowStart(d *schema.ResourceData) *timeofday.TimeOfDay {
	out := &timeofday.TimeOfDay{}

	if v, ok := d.GetOk("config.0.backup_window_start.0.hours"); ok {
		out.Hours = int32(v.(int))
	}

	if v, ok := d.GetOk("config.0.backup_window_start.0.minutes"); ok {
		out.Minutes = int32(v.(int))
	}

	return out
}

func expandPGAccess(d *schema.ResourceData) *postgresql.Access {
	out := &postgresql.Access{}

	if v, ok := d.GetOk("config.0.access.0.data_lens"); ok {
		out.DataLens = v.(bool)
	}

	return out
}

func pgUsersDiff(currUsers []*postgresql.User, targetUsers []*postgresql.UserSpec) ([]string, []*postgresql.UserSpec) {
	m := map[string]bool{}
	toDelete := map[string]bool{}
	toAdd := []*postgresql.UserSpec{}

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

type IndexedUserSpec struct {
	index int
	user  *postgresql.UserSpec
}

func pgChangedUsers(oldSpecs []interface{}, newSpecs []interface{}, withNewPermissions bool) ([]IndexedUserSpec, error) {
	out := make([]IndexedUserSpec, 0)

	m := map[string]*postgresql.UserSpec{}
	permissions := map[string]interface{}{}
	for _, spec := range oldSpecs {
		mapOldUser := spec.(map[string]interface{})
		user, err := expandPGUser(mapOldUser)
		if err != nil {
			return nil, err
		}
		m[user.Name] = user
		permissions[user.Name] = mapOldUser["permission"]
	}

	for i, spec := range newSpecs {
		mapUser := spec.(map[string]interface{})
		newUser, err := expandPGUser(mapUser)
		if err != nil {
			return nil, err
		}
		if oldUser, ok := m[newUser.Name]; ok {
			if !withNewPermissions {
				newUser.Permissions = oldUser.Permissions
			}
			if !reflect.DeepEqual(newUser, oldUser) {
				out = append(out, IndexedUserSpec{i, newUser})
			}
		} else if withNewPermissions {
			out = append(out, IndexedUserSpec{i, newUser})
		}
	}

	return out, nil
}

func pgDatabasesDiff(currDBs []*postgresql.Database, targetDBs []*postgresql.DatabaseSpec) ([]string, []*postgresql.DatabaseSpec) {
	m := map[string]bool{}
	toAdd := []*postgresql.DatabaseSpec{}
	toDelete := map[string]bool{}
	for _, db := range currDBs {
		toDelete[db.Name] = true
		m[db.Name] = true
	}

	for _, db := range targetDBs {
		delete(toDelete, db.Name)
		if _, ok := m[db.Name]; !ok {
			toAdd = append(toAdd, db)
		}
	}

	toDel := []string{}
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func pgChangedDatabases(oldSpecs []interface{}, newSpecs []interface{}) ([]*postgresql.DatabaseSpec, error) {
	out := []*postgresql.DatabaseSpec{}

	m := map[string]*postgresql.DatabaseSpec{}
	for _, spec := range oldSpecs {
		db, err := expandPGDatabase(spec.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		m[db.Name] = db
	}

	for _, spec := range newSpecs {
		db, err := expandPGDatabase(spec.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		if oldDB, ok := m[db.Name]; ok {
			if !reflect.DeepEqual(db, oldDB) {
				out = append(out, db)
			}
		}
	}

	return out, nil
}

func pgHostsDiff(currHosts []*postgresql.Host, targetHosts []*PostgreSQLHostSpec) ([]string, []*postgresql.HostSpec) {
	m := map[string]*PostgreSQLHostSpec{}

	toAdd := []*postgresql.HostSpec{}
	for _, h := range targetHosts {
		if !h.HasComputedFqdn {
			toAdd = append(toAdd, h.HostSpec)
		} else {
			m[h.Fqdn] = h
		}
	}

	toDelete := []string{}
	for _, h := range currHosts {
		_, ok := m[h.Name]
		if !ok {
			toDelete = append(toDelete, h.Name)
		}
	}

	return toDelete, toAdd
}

func parsePostgreSQLEnv(e string) (postgresql.Cluster_Environment, error) {
	v, ok := postgresql.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(postgresql.Cluster_Environment_value)), e)
	}

	return postgresql.Cluster_Environment(v), nil
}

func parsePostgreSQLPoolingMode(s string) (postgresql.ConnectionPoolerConfig_PoolingMode, error) {
	v, ok := postgresql.ConnectionPoolerConfig_PoolingMode_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'pooling_mode' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(postgresql.ConnectionPoolerConfig_PoolingMode_value)), s)
	}

	return postgresql.ConnectionPoolerConfig_PoolingMode(v), nil
}
