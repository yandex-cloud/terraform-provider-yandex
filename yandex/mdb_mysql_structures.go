package yandex

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/objx"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
	"google.golang.org/genproto/googleapis/type/timeofday"
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
		var list []interface{}
		list, ok = v.([]interface{}) // old schema used List instead of Set. Keep List as fallback
		if !ok {
			list = v.(*schema.Set).List()
		}
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

type compareMySQLHostsInfoResult struct {
	hostsInfo        map[string]*myHostInfo // fqdn -> *myHostInfo
	createHostsInfo  []*myHostInfo          // hosts to be created
	haveHostWithName bool
	hierarchyExists  bool
}

type myHostInfo struct {
	name string
	fqdn string

	zone     string
	subnetID string

	oldAssignPublicIP        bool
	oldReplicationSource     string
	oldReplicationSourceName string
	oldPriority              int64
	oldBackupPriority        int64

	newAssignPublicIP        bool
	newReplicationSource     string
	newReplicationSourceName string
	newPriority              int64
	newBackupPriority        int64

	// inTargetSet is true when host is present in target set (and shouldn't be removed)
	inTargetSet bool

	rowNumber int
}

type MySQLHostSpec struct {
	HostSpec              *mysql.HostSpec
	Fqdn                  string
	Name                  string
	ReplicationSourceName string
	Priority              int64
	BackupPriority        int64
}

func expandMysqlHostSpec(d *schema.ResourceData) ([]*mysql.HostSpec, error) {
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

func expandEnrichedMySQLHostSpec(d *schema.ResourceData) ([]*MySQLHostSpec, error) {
	var result []*MySQLHostSpec
	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host, err := expandEnrichedMySQLHost(config)
		if err != nil {
			return nil, err
		}
		result = append(result, host)
	}

	return result, nil
}

func expandMysqlHost(config map[string]interface{}) (*mysql.HostSpec, error) {
	hostSpec := &mysql.HostSpec{}
	if v, ok := config["zone"]; ok {
		hostSpec.ZoneId = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		hostSpec.SubnetId = v.(string)
	}

	if v, ok := config["assign_public_ip"]; ok {
		hostSpec.AssignPublicIp = v.(bool)
	}

	if v, ok := config["backup_priority"]; ok {
		hostSpec.BackupPriority = int64(v.(int))
	}

	if v, ok := config["priority"]; ok {
		hostSpec.Priority = int64(v.(int))
	}

	return hostSpec, nil
}

func expandEnrichedMySQLHost(config map[string]interface{}) (*MySQLHostSpec, error) {
	hostSpec, err := expandMysqlHost(config)
	if err != nil {
		return nil, err
	}
	mysqlHostSpec := &MySQLHostSpec{HostSpec: hostSpec}
	if v, ok := config["fqdn"]; ok && v.(string) != "" {
		mysqlHostSpec.Fqdn = v.(string)
	}
	if v, ok := config["name"]; ok {
		mysqlHostSpec.Name = v.(string)
	}
	if v, ok := config["replication_source_name"]; ok {
		mysqlHostSpec.ReplicationSourceName = v.(string)
	}
	if v, ok := config["priority"]; ok {
		mysqlHostSpec.Priority = int64(v.(int))
	}
	if v, ok := config["backup_priority"]; ok {
		mysqlHostSpec.BackupPriority = int64(v.(int))
	}
	return mysqlHostSpec, nil
}

func validateMysqlReplicationReferences(targetHosts []*MySQLHostSpec) error {
	// Names are unique:
	names := map[string]bool{}
	for _, host := range targetHosts {
		if host.Name != "" {
			if _, ok := names[host.Name]; ok {
				return fmt.Errorf("duplicate host names '%s' in resource_yandex_mdb_mysql_cluster", host.Name)
			}
			names[host.Name] = true
		}
	}

	if len(names) != 0 && len(names) != len(targetHosts) {
		return fmt.Errorf("all or none hosts should have names")
	}

	// ReplicationSourceName refers to existing names:
	for _, host := range targetHosts {
		if host.ReplicationSourceName != "" {
			if !names[host.ReplicationSourceName] {
				return fmt.Errorf("replication_source_name '%s' for host '%s' doesn't exists", host.ReplicationSourceName, host.Name)
			}
		}
	}

	// self-replication is not allowed:
	for _, host := range targetHosts {
		if host.Name != "" && host.Name == host.ReplicationSourceName {
			return fmt.Errorf("host with name '%s' refers to itself as to replication source", host.Name)
		}
	}

	// all hosts are reachable from HA-group (no loops)
	if len(names) != 0 {
		visited := map[string]bool{}
		for _, host := range targetHosts {
			if host.ReplicationSourceName == "" { // HA-nodes
				visited[host.Name] = true
			}
		}
		if len(visited) == 0 {
			return fmt.Errorf("there should be at least one HA-node in cluster")
		}
		for {
			if len(visited) == len(targetHosts) {
				break // Ok. all hosts are reachable
			}

			numVisited := len(visited)
			for _, host := range targetHosts {
				if host.ReplicationSourceName != "" && visited[host.ReplicationSourceName] {
					visited[host.Name] = true
				}
			}
			if len(visited) == numVisited {
				unreachableHosts := make([]string, 0)
				for _, host := range targetHosts {
					if !visited[host.Name] {
						unreachableHosts = append(unreachableHosts, host.Name)
					}
				}
				return fmt.Errorf("there is no replication chain from HA-hosts to following hosts: '%s' (probably, there is a loop in replication_source chain)", strings.Join(unreachableHosts, ", "))
			}
		}
	}

	return nil
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

func addMySQLHost(ctx context.Context, config *Config, d *schema.ResourceData, host *mysql.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().Cluster().AddHosts(ctx, &mysql.AddClusterHostsRequest{
			ClusterId: d.Id(),
			HostSpecs: []*mysql.HostSpec{host},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create host for MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating host for MySQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating host for MySQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func updateMySQLHost(ctx context.Context, config *Config, d *schema.ResourceData, host *mysql.UpdateHostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().Cluster().UpdateHosts(ctx, &mysql.UpdateClusterHostsRequest{
			ClusterId:       d.Id(),
			UpdateHostSpecs: []*mysql.UpdateHostSpec{host},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update host for MySQL Cluster %q - host %v: %s", d.Id(), host.HostName, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating host for MySQL Cluster %q - host %v: %s", d.Id(), host.HostName, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating host for MySQL Cluster %q - host %v failed: %s", d.Id(), host.HostName, err)
	}

	return nil
}

func loadNewMySQLHostsInfo(newHosts []interface{}) (hostsInfo []*myHostInfo, err error) {
	hostsInfo = make([]*myHostInfo, 0)
	for i, hostNewInfo := range newHosts {
		hni := objx.New(hostNewInfo)
		if hni == nil {
			return nil, fmt.Errorf("MySQL.host: failed to read hosts info %v", hostsInfo)
		}
		hostsInfo = append(hostsInfo, &myHostInfo{
			name:                     hni.Get("name").Str(),
			zone:                     hni.Get("zone").Str(),
			subnetID:                 hni.Get("subnet_id").Str(),
			newAssignPublicIP:        hni.Get("assign_public_ip").Bool(),
			rowNumber:                i,
			newReplicationSourceName: hni.Get("replication_source_name").Str(),
			// because hni.Get("priority").Int64() results to 0
			newPriority:       int64(hni.Get("priority").Int()),
			newBackupPriority: int64(hni.Get("backup_priority").Int()),
		})

	}
	return hostsInfo, nil
}

func validateNewMySQLHostsInfo(newHostsInfo []*myHostInfo, isUpdate bool) (haveHostWithName bool, err error) {
	uniqueNames := make(map[string]struct{})
	haveHostWithoutName := false

	for _, nhi := range newHostsInfo {
		name := nhi.name
		if name == "" {
			haveHostWithoutName = true
		} else {
			haveHostWithName = true
			if _, ok := uniqueNames[name]; ok && isUpdate {
				return haveHostWithName, fmt.Errorf("MySQL.host: name is duplicate %v", name)
			}
			uniqueNames[name] = struct{}{}
		}
	}

	if haveHostWithName && haveHostWithoutName && isUpdate {
		return haveHostWithName, fmt.Errorf("names should be set for all hosts or unset for all host")
	}

	return haveHostWithName, nil
}

func compareMySQLNamedHostInfo(existsHostInfo *myHostInfo, newHostInfo *myHostInfo, nameToHost map[string]string) int {
	if existsHostInfo.name == newHostInfo.name {
		return 10
	}
	if existsHostInfo.name != "" {
		return 0
	}

	if existsHostInfo.zone != newHostInfo.zone ||
		existsHostInfo.subnetID != newHostInfo.subnetID && newHostInfo.subnetID != "" {
		return 0
	}

	compareWeight := 1

	if newHostInfo.newReplicationSourceName != "" {
		if fqdn, ok := nameToHost[newHostInfo.newReplicationSourceName]; ok && existsHostInfo.oldReplicationSource == fqdn {
			compareWeight += 4
		}
	}

	if existsHostInfo.oldAssignPublicIP == newHostInfo.newAssignPublicIP {
		compareWeight++
	}

	if existsHostInfo.oldBackupPriority == newHostInfo.newBackupPriority {
		compareWeight++
	}

	if existsHostInfo.oldPriority == newHostInfo.newPriority {
		compareWeight++
	}

	return compareWeight
}

func matchesMySQLNoNamedHostInfo(existsHostInfo *myHostInfo, newHostInfo *myHostInfo) bool {
	if existsHostInfo.zone != newHostInfo.zone ||
		existsHostInfo.subnetID != newHostInfo.subnetID && newHostInfo.subnetID != "" {
		return false
	}

	if existsHostInfo.oldAssignPublicIP != newHostInfo.newAssignPublicIP {
		return false
	}

	if existsHostInfo.oldBackupPriority != newHostInfo.newBackupPriority {
		return false
	}

	if existsHostInfo.oldPriority != newHostInfo.newPriority {
		return false
	}

	return true
}

func compareMySQLNamedHostsInfoWeight(existsHostsInfo map[string]*myHostInfo, newHostsInfo []*myHostInfo, compareMap map[int]string) int {
	weight := 0

	nameToHost := make(map[string]string)
	for row, fqdn := range compareMap {
		name := newHostsInfo[row].name
		if name != "" {
			nameToHost[name] = fqdn
		}
	}

	for row, fqdn := range compareMap {
		weightStep := compareMySQLNamedHostInfo(existsHostsInfo[fqdn], newHostsInfo[row], nameToHost)
		if weightStep == 0 {
			return 0
		}

		weight += weightStep
	}

	return weight
}

type mysqlHostMapper struct {
	// static data:
	existingHostsInfo map[string]*myHostInfo
	targetHostsInfo   []*myHostInfo
	nameToHost        map[string]string
}

// Recursively generate all reasonable matching existing->target hosts configurations
// (some values may have no matching counterparty in target hosts)
// Calls cb function for every possible combination. Note: cb argument is mutable
func (mapper *mysqlHostMapper) findBestMatch(state map[int]string, itm int, cb func(map[int]string)) {
	if len(mapper.targetHostsInfo) <= itm {
		cb(state)
		return
	}

	newHostInfo := mapper.targetHostsInfo[itm]
	// when terraform already knows 'name' to 'fqdn' mapping - we 100% sure that this is good match
	if fqdn, ok := mapper.nameToHost[newHostInfo.name]; ok {
		state[itm] = fqdn
		mapper.findBestMatch(state, itm+1, cb)
		return
	}

	// Case: assume that no matching existing host found
	delete(state, itm)
	mapper.findBestMatch(copyMapIntString(state), itm+1, cb)

outer:
	for fqdn, existHostInfo := range mapper.existingHostsInfo {
		for i := 0; i < itm; i++ { // no duplicates allowed
			if state[i] == fqdn {
				continue outer
			}
		}

		// some pairs couldn't be matched - skip such combinations
		weight := compareMySQLNamedHostInfo(existHostInfo, newHostInfo, mapper.nameToHost)
		if weight == 0 {
			continue
		}

		state[itm] = fqdn
		mapper.findBestMatch(state, itm+1, cb)
	}
}

// row idx (in newHostsInfo) -> FQDN
func compareMySQLNamedHostsInfo(existsHostsInfo map[string]*myHostInfo, newHostsInfo []*myHostInfo) map[int]string {
	nameToHost := make(map[string]string)
	for fqdn, hi := range existsHostsInfo {
		if hi.name != "" {
			nameToHost[hi.name] = fqdn
		}
	}

	mysqlHostMapper := mysqlHostMapper{
		existingHostsInfo: existsHostsInfo,
		targetHostsInfo:   newHostsInfo,
		nameToHost:        nameToHost,
	}
	// Find best existingHostsInfo to targetHostsInfo match:
	weight := 0
	compareMap := make(map[int]string)
	mysqlHostMapper.findBestMatch(map[int]string{}, 0, func(candidate map[int]string) {
		stepWeight := compareMySQLNamedHostsInfoWeight(existsHostsInfo, newHostsInfo, candidate)
		if stepWeight > weight {
			weight = stepWeight
			compareMap = copyMapIntString(candidate)
		}
	})
	return compareMap
}

// row idx (in newHostsInfo) -> FQDN
func compareMySQLNoNamedHostsInfo(existingHostsInfo map[string]*myHostInfo, targetHostsInfo []*myHostInfo) map[int]string {
	compareMap := make(map[int]string)
	visitedHostNames := make(map[string]struct{})

	for i, targetHostInfo := range targetHostsInfo {
		for _, existingHostInfo := range existingHostsInfo {
			if _, ok := visitedHostNames[existingHostInfo.fqdn]; ok {
				continue
			}
			if matchesMySQLNoNamedHostInfo(existingHostInfo, targetHostInfo) {
				visitedHostNames[existingHostInfo.fqdn] = struct{}{}
				compareMap[i] = existingHostInfo.fqdn
				break
			}
		}
	}

	return compareMap
}

func loadExistingMySQLHostsInfo(currentHosts []*mysql.Host, oldHosts []interface{}) (map[string]*myHostInfo, error) {
	hostsInfo := make(map[string]*myHostInfo)

	for i, h := range currentHosts {
		// Note: mysql.Host.Name is the FQDN of the host
		hostsInfo[h.Name] = &myHostInfo{
			fqdn:                 h.Name,
			zone:                 h.ZoneId,
			subnetID:             h.SubnetId,
			oldAssignPublicIP:    h.AssignPublicIp,
			oldReplicationSource: h.ReplicationSource,
			oldPriority:          h.Priority,
			oldBackupPriority:    h.BackupPriority,

			rowNumber: i,
		}
	}

	for _, hostOldInfo := range oldHosts {
		hoi := objx.New(hostOldInfo)
		if hoi == nil {
			return nil, fmt.Errorf("MySQL.host: failed to read hosts info %v", hostsInfo)
		}

		if !hoi.Has("fqdn") || !hoi.Has("name") {
			continue
		}
		fqdn := hoi.Get("fqdn").Str()
		name := hoi.Get("name").Str()

		if hi, ok := hostsInfo[fqdn]; ok {
			hi.name = name
		}
	}

	return hostsInfo, nil
}

func compareMySQLHostsInfo(d *schema.ResourceData, currentHosts []*mysql.Host, isUpdate bool) (compareMySQLHostsInfoResult, error) {

	result := compareMySQLHostsInfoResult{}

	oldHosts, newHosts := d.GetChange("host")

	// actual hosts configuration (enriched with 'name', when available): fqdn -> *myHostInfo
	existingHostsInfo, err := loadExistingMySQLHostsInfo(currentHosts, oldHosts.([]interface{}))
	if err != nil {
		return result, err
	}

	// expected hosts configuration: []*myHostInfo
	newHostsInfo, err := loadNewMySQLHostsInfo(newHosts.([]interface{}))
	if err != nil {
		return result, err
	}

	result.haveHostWithName, err = validateNewMySQLHostsInfo(newHostsInfo, isUpdate)
	if err != nil {
		return result, err
	}

	nameToHost := make(map[string]string)
	for fqdn, hi := range existingHostsInfo {
		if hi.name != "" {
			nameToHost[hi.name] = fqdn
		}
	}

	if result.haveHostWithName {
		// find best mapping from existingHostsInfo to newHostsInfo
		compareMap := compareMySQLNamedHostsInfo(existingHostsInfo, newHostsInfo)

		hostToName := make(map[string]string)
		for i, fqdn := range compareMap {
			hostToName[fqdn] = newHostsInfo[i].name
			log.Printf("[DEBUG] match [%d]: %s -> %s", i, newHostsInfo[i].name, fqdn)
		}

		result.hostsInfo = existingHostsInfo
		for i, newHostInfo := range newHostsInfo {
			if existHostFqdn, ok := compareMap[i]; ok {
				existHostInfo := existingHostsInfo[existHostFqdn]
				existHostInfo.name = newHostInfo.name
				existHostInfo.rowNumber = newHostInfo.rowNumber
				existHostInfo.newReplicationSourceName = newHostInfo.newReplicationSourceName
				existHostInfo.newAssignPublicIP = newHostInfo.newAssignPublicIP
				existHostInfo.newPriority = newHostInfo.newPriority
				existHostInfo.newBackupPriority = newHostInfo.newBackupPriority
				existHostInfo.inTargetSet = true
				nameToHost[existHostInfo.name] = existHostInfo.fqdn
			}
		}

		for _, existHostInfo := range existingHostsInfo {
			if existHostInfo.newReplicationSourceName == "" {
				existHostInfo.newReplicationSource = ""
			} else if fqdn, ok := nameToHost[existHostInfo.newReplicationSourceName]; ok {
				existHostInfo.newReplicationSource = fqdn
			} else {
				existHostInfo.newReplicationSource = existHostInfo.oldReplicationSource
			}

			if existHostInfo.oldReplicationSource == "" {
				existHostInfo.oldReplicationSourceName = ""
			} else if name, ok := hostToName[existHostInfo.oldReplicationSource]; ok {
				existHostInfo.oldReplicationSourceName = name
			}
		}

		createHostsInfoPrepare := make([]*myHostInfo, 0)
		for i, newHostInfo := range newHostsInfo {
			if _, ok := compareMap[i]; !ok { // for all hosts for which we don't have mapping:
				if newHostInfo.newReplicationSourceName == "" {
					createHostsInfoPrepare = append(createHostsInfoPrepare, newHostInfo)
				} else {
					fqdn, ok := nameToHost[newHostInfo.newReplicationSourceName] // resolve cascade Name to FQDN
					if ok {
						newHostInfo.newReplicationSource = fqdn
						createHostsInfoPrepare = append(createHostsInfoPrepare, newHostInfo)
					} else {
						result.hierarchyExists = true
					}
				}
			}
		}

		result.createHostsInfo = createHostsInfoPrepare
		return result, nil

	}

	createHostsInfoPrepare := make([]*myHostInfo, 0)
	compareMap := compareMySQLNoNamedHostsInfo(existingHostsInfo, newHostsInfo)
	for i, newHostInfo := range newHostsInfo {
		if existHostFqdn, ok := compareMap[i]; ok {
			existHostInfo := existingHostsInfo[existHostFqdn]
			existHostInfo.rowNumber = newHostsInfo[i].rowNumber
			existHostInfo.inTargetSet = true
		} else {
			createHostsInfoPrepare = append(createHostsInfoPrepare, newHostInfo)
		}
	}

	result.hostsInfo = existingHostsInfo
	result.createHostsInfo = createHostsInfoPrepare
	return result, nil
}

func flattenMysqlHosts(d *schema.ResourceData, hs []*mysql.Host, isDataSource bool) ([]map[string]interface{}, error) {
	// read operation should return hosts in the same order, as defined in terraform file (otherwise Terraform
	// will think that some diff exists and should be fixed)
	// so, we should sort retrieved hosts:
	compareHostsInfo, err := compareMySQLHostsInfo(d, hs, false)
	if err != nil {
		return nil, err
	}

	hosts := flattenMySQLHostsFromHostInfo(compareHostsInfo.hostsInfo, isDataSource)
	return hosts, nil
}

func flattenMySQLHostsFromHostInfo(hostsInfo map[string]*myHostInfo, isDataSource bool) []map[string]interface{} {
	orderedHostsInfo := make([]*myHostInfo, 0, len(hostsInfo))
	for _, hostInfo := range hostsInfo {
		orderedHostsInfo = append(orderedHostsInfo, hostInfo)
	}
	sort.Slice(orderedHostsInfo, func(i, j int) bool {
		if orderedHostsInfo[i].inTargetSet == orderedHostsInfo[j].inTargetSet {
			return orderedHostsInfo[i].rowNumber < orderedHostsInfo[j].rowNumber
		}
		return orderedHostsInfo[i].inTargetSet
	})

	hosts := []map[string]interface{}{}
	for _, hostInfo := range orderedHostsInfo {
		m := map[string]interface{}{}
		m["zone"] = hostInfo.zone
		m["subnet_id"] = hostInfo.subnetID
		m["assign_public_ip"] = hostInfo.oldAssignPublicIP
		m["fqdn"] = hostInfo.fqdn
		m["replication_source"] = hostInfo.oldReplicationSource
		m["priority"] = hostInfo.oldPriority
		m["backup_priority"] = hostInfo.oldBackupPriority
		if !isDataSource {
			m["name"] = hostInfo.name
			m["replication_source_name"] = hostInfo.oldReplicationSourceName
		}

		hosts = append(hosts, m)
	}

	return hosts
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

func flattenMysqlMaintenanceWindow(mw *mysql.MaintenanceWindow) ([]interface{}, error) {
	maintenanceWindow := map[string]interface{}{}
	if mw != nil {
		switch p := mw.GetPolicy().(type) {
		case *mysql.MaintenanceWindow_Anytime:
			maintenanceWindow["type"] = "ANYTIME"
			// do nothing
		case *mysql.MaintenanceWindow_WeeklyMaintenanceWindow:
			maintenanceWindow["type"] = "WEEKLY"
			maintenanceWindow["hour"] = p.WeeklyMaintenanceWindow.Hour
			maintenanceWindow["day"] = mysql.WeeklyMaintenanceWindow_WeekDay_name[int32(p.WeeklyMaintenanceWindow.GetDay())]
		default:
			return nil, fmt.Errorf("unsupported Mysql maintenance policy type")
		}
	}

	return []interface{}{maintenanceWindow}, nil
}

func expandMySQLMaintenanceWindow(d *schema.ResourceData) (*mysql.MaintenanceWindow, error) {
	if _, ok := d.GetOkExists("maintenance_window"); !ok {
		return nil, nil
	}

	out := &mysql.MaintenanceWindow{}
	typeMW, _ := d.GetOk("maintenance_window.0.type")
	if typeMW == "ANYTIME" {
		if hour, ok := d.GetOk("maintenance_window.0.hour"); ok && hour != "" {
			return nil, fmt.Errorf("hour should be not set, when using ANYTIME")
		}
		if day, ok := d.GetOk("maintenance_window.0.day"); ok && day != "" {
			return nil, fmt.Errorf("day should be not set, when using ANYTIME")
		}
		out.Policy = &mysql.MaintenanceWindow_Anytime{
			Anytime: &mysql.AnytimeMaintenanceWindow{},
		}
	} else if typeMW == "WEEKLY" {
		hour := d.Get("maintenance_window.0.hour").(int)
		dayString := d.Get("maintenance_window.0.day").(string)

		day, ok := mysql.WeeklyMaintenanceWindow_WeekDay_value[dayString]
		if !ok || day == 0 {
			return nil, fmt.Errorf(`day value should be one of ("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN")`)
		}

		out.Policy = &mysql.MaintenanceWindow_WeeklyMaintenanceWindow{
			WeeklyMaintenanceWindow: &mysql.WeeklyMaintenanceWindow{
				Hour: int64(hour),
				Day:  mysql.WeeklyMaintenanceWindow_WeekDay(day),
			},
		}
	} else {
		return nil, fmt.Errorf("maintenance_window.0.type should be ANYTIME or WEEKLY")
	}

	return out, nil
}

func mysqlMaintenanceWindowSchemaValidateFunc(v interface{}, k string) (s []string, es []error) {
	dayString := v.(string)
	day, ok := mysql.WeeklyMaintenanceWindow_WeekDay_value[dayString]
	if !ok || day == 0 {
		es = append(es, fmt.Errorf(`expected %s value should be one of ("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"). Current value is %v`, k, v))
		return
	}

	return
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

		settings, err := flattenResourceGenerateMapS(cf.MysqlConfig_8_0.UserConfig, false, mdbMySQLSettingsFieldsInfo, false, true, nil)
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
		settings, err := flattenResourceGenerateMapS(cf.MysqlConfig_5_7.UserConfig, false, mdbMySQLSettingsFieldsInfo, false, true, nil)
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

func expandMyPerformanceDiagnostics(d *schema.ResourceData) *mysql.PerformanceDiagnostics {

	if _, ok := d.GetOkExists("performance_diagnostics"); !ok {
		return nil
	}

	out := &mysql.PerformanceDiagnostics{}

	if v, ok := d.GetOk("performance_diagnostics.0.enabled"); ok {
		out.Enabled = v.(bool)
	}

	if v, ok := d.GetOk("performance_diagnostics.0.sessions_sampling_interval"); ok {
		out.SessionsSamplingInterval = int64(v.(int))
	}

	if v, ok := d.GetOk("performance_diagnostics.0.statements_sampling_interval"); ok {
		out.StatementsSamplingInterval = int64(v.(int))
	}

	return out
}

func flattenMyPerformanceDiagnostics(p *mysql.PerformanceDiagnostics) ([]interface{}, error) {
	if p == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["enabled"] = p.Enabled
	out["sessions_sampling_interval"] = int(p.SessionsSamplingInterval)
	out["statements_sampling_interval"] = int(p.StatementsSamplingInterval)

	return []interface{}{out}, nil
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
