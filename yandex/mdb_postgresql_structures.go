package yandex

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/stretchr/objx"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	timeofday "google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
)

type PostgreSQLHostSpec struct {
	HostSpec *postgresql.HostSpec
	Fqdn     string
}

func flattenPGClusterConfig(c *postgresql.ClusterConfig, d *schema.ResourceData) ([]interface{}, error) {
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

	performanceDiagnostics, err := flattenPGPerformanceDiagnostics(c.PerformanceDiagnostics)
	if err != nil {
		return nil, err
	}

	access, err := flattenPGAccess(c.Access)
	if err != nil {
		return nil, err
	}

	settings, err := flattenPGSettings(c, d)
	if err != nil {
		return nil, err
	}

	out := map[string]interface{}{}
	out["autofailover"] = c.GetAutofailover().GetValue()
	out["version"] = c.Version
	out["pooler_config"] = poolerConf
	out["resources"] = resources
	out["backup_window_start"] = backupWindowStart
	out["backup_retain_period_days"] = c.BackupRetainPeriodDays.GetValue()
	out["performance_diagnostics"] = performanceDiagnostics
	out["access"] = access
	out["postgresql_config"] = settings

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

func flattenPGPerformanceDiagnostics(p *postgresql.PerformanceDiagnostics) ([]interface{}, error) {
	if p == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["enabled"] = p.Enabled
	out["sessions_sampling_interval"] = int(p.SessionsSamplingInterval)
	out["statements_sampling_interval"] = int(p.StatementsSamplingInterval)

	return []interface{}{out}, nil
}

func flattenPGSettingsSPL(settings map[string]string, d *schema.ResourceData) map[string]string {
	spl, ok := d.GetOkExists("config.0.postgresql_config.shared_preload_libraries")
	if ok {
		if settings == nil {
			settings = make(map[string]string)
		}
		if _, exists := settings["shared_preload_libraries"]; !exists {
			settings["shared_preload_libraries"] = spl.(string)
		}
	}

	return settings
}

func flattenPGSettings(c *postgresql.ClusterConfig, d *schema.ResourceData) (map[string]string, error) {
	// TODO refactor it
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_14); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_14.UserConfig, false, mdbPGSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		settings = flattenPGSettingsSPL(settings, d)

		return settings, err
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_13); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_13.UserConfig, false, mdbPGSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		settings = flattenPGSettingsSPL(settings, d)

		return settings, err
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_12); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_12.UserConfig, false, mdbPGSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		settings = flattenPGSettingsSPL(settings, d)

		return settings, err
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_12_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_12_1C.UserConfig, false, mdbPGSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		settings = flattenPGSettingsSPL(settings, d)
		return settings, err
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_11); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_11.UserConfig, false, mdbPGSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		settings = flattenPGSettingsSPL(settings, d)
		return settings, err
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_11_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_11_1C.UserConfig, false, mdbPGSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		settings = flattenPGSettingsSPL(settings, d)
		return settings, err
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_10); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_10.UserConfig, false, mdbPGSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		settings = flattenPGSettingsSPL(settings, d)
		return settings, err
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_10_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_10_1C.UserConfig, false, mdbPGSettingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}

		settings = flattenPGSettingsSPL(settings, d)
		return settings, err
	}

	return nil, nil
}

func flattenPGAccess(a *postgresql.Access) ([]interface{}, error) {
	if a == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["data_lens"] = a.DataLens
	out["web_sql"] = a.WebSql
	out["serverless"] = a.Serverless

	return []interface{}{out}, nil
}

func flattenPGUsers(us []*postgresql.User, passwords map[string]string,
	fieldsInfo *objectFieldsInfo) ([]map[string]interface{}, error) {

	out := make([]map[string]interface{}, 0)

	for _, u := range us {
		knownDefault := map[string]struct{}{
			"log_min_duration_statement": {},
		}
		ou, err := flattenPGUser(u, fieldsInfo, knownDefault)
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

func flattenPGUser(u *postgresql.User,
	fieldsInfo *objectFieldsInfo, knownDefault map[string]struct{}) (map[string]interface{}, error) {
	settings, err := flattenResourceGenerateMapS(u.Settings, false, fieldsInfo, false, true, knownDefault)
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{}
	m["name"] = u.Name
	m["login"] = u.GetLogin().GetValue()
	m["permission"] = flattenPGUserPermissions(u.Permissions)
	m["grants"] = u.Grants
	m["conn_limit"] = u.ConnLimit
	m["settings"] = settings

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

func flattenPGUserPermissions(ps []*postgresql.Permission) *schema.Set {
	out := schema.NewSet(pgUserPermissionHash, nil)

	for _, p := range ps {
		op := map[string]interface{}{
			"database_name": p.DatabaseName,
		}

		out.Add(op)
	}

	return out
}

type pgHostInfo struct {
	name string
	fqdn string

	zone     string
	subnetID string

	role postgresql.Host_Role

	oldAssignPublicIP        bool
	oldPriority              int
	oldReplicationSource     string
	oldReplicationSourceName string

	newAssignPublicIP        bool
	newPriority              int
	newReplicationSource     string
	newReplicationSourceName string

	inTargetSet bool

	rowNumber int
}

func loadNewPGHostsInfo(newHosts []interface{}) ([]*pgHostInfo, error) {
	hostsInfo := make([]*pgHostInfo, 0)
	for i, newHostInfo := range newHosts {
		hni := objx.New(newHostInfo)
		if hni == nil {
			return nil, fmt.Errorf("PostgreSQL.host: failed to read hosts info %v", hostsInfo)
		}
		hostsInfo = append(hostsInfo, &pgHostInfo{
			name:                     hni.Get("name").Str(),
			zone:                     hni.Get("zone").Str(),
			subnetID:                 hni.Get("subnet_id").Str(),
			newAssignPublicIP:        hni.Get("assign_public_ip").Bool(),
			newReplicationSourceName: hni.Get("replication_source_name").Str(),
			newPriority:              hni.Get("priority").Int(),
			rowNumber:                i,
			inTargetSet:              true,
		})

	}

	return hostsInfo, nil
}

func comparePGNamedHostInfo(existsHostInfo *pgHostInfo, newHostInfo *pgHostInfo, currentNameToHost map[string]string, hostMasterName string) int {
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

	if hostMasterName != "" &&
		(existsHostInfo.role == postgresql.Host_MASTER && hostMasterName == newHostInfo.name ||
			existsHostInfo.role != postgresql.Host_MASTER && hostMasterName != newHostInfo.name) {
		compareWeight += 2
	}

	if newHostInfo.newReplicationSourceName != "" {
		if fqdn, ok := currentNameToHost[newHostInfo.newReplicationSourceName]; ok && existsHostInfo.oldReplicationSource == fqdn {
			compareWeight += 4
		}
	}

	if existsHostInfo.oldPriority == newHostInfo.newPriority {
		compareWeight++
	}

	if existsHostInfo.oldAssignPublicIP == newHostInfo.newAssignPublicIP {
		compareWeight++
	}

	return compareWeight
}

func matchesPGNoNamedHostInfo(existsHostInfo *pgHostInfo, newHostInfo *pgHostInfo) bool {
	if existsHostInfo.zone != newHostInfo.zone ||
		existsHostInfo.subnetID != newHostInfo.subnetID && newHostInfo.subnetID != "" {
		return false
	}

	if existsHostInfo.oldAssignPublicIP != newHostInfo.newAssignPublicIP {
		return false
	}

	if existsHostInfo.oldPriority != newHostInfo.newPriority {
		return false
	}

	return true
}

func copyMapStringString(source map[string]string) map[string]string {
	res := make(map[string]string)
	for k, v := range source {
		res[k] = v
	}
	return res
}

func copyMapIntString(source map[int]string) map[int]string {
	res := make(map[int]string)
	for k, v := range source {
		res[k] = v
	}
	return res
}

func comparePGNamedHostsInfoWeight(existsHostsInfo map[string]*pgHostInfo, newHostsInfo []*pgHostInfo, compareResult pgCompareHostNameResult, hostMasterName string) int {
	weight := 0

	for i, fqdn := range compareResult.compareMap {
		weightStep := comparePGNamedHostInfo(existsHostsInfo[fqdn], newHostsInfo[i], compareResult.nameToHost, hostMasterName)
		if weightStep == 0 {
			return 0
		}

		weight += weightStep
	}

	return weight
}

type pgCompareHostNameResult struct {
	nameToHost map[string]string
	compareMap map[int]string
	hostToName map[string]string
}

func generatePGNamedHostsInfoMaps(existsHostsInfo map[string]*pgHostInfo, newHostsInfo []*pgHostInfo, nameToHost map[string]string, hostToName map[string]string, compareMap map[int]string, itm int, hostMasterName string) (compareResults []pgCompareHostNameResult) {
	compareResults = make([]pgCompareHostNameResult, 0)

	if len(newHostsInfo) <= itm {
		compareResults = append(compareResults, pgCompareHostNameResult{
			nameToHost: nameToHost,
			compareMap: compareMap,
			hostToName: hostToName,
		})
		return compareResults
	}

	newHostInfo := newHostsInfo[itm]

	if fqdn, ok := nameToHost[newHostInfo.name]; ok {
		compareMap[itm] = fqdn
		hostToName[fqdn] = newHostInfo.name
		return generatePGNamedHostsInfoMaps(existsHostsInfo, newHostsInfo, nameToHost, hostToName, compareMap, itm+1, hostMasterName)
	}

	compareResults = append(compareResults, generatePGNamedHostsInfoMaps(existsHostsInfo, newHostsInfo, copyMapStringString(nameToHost), copyMapStringString(hostToName), copyMapIntString(compareMap), itm+1, hostMasterName)...)

	for fqdn, existHostInfo := range existsHostsInfo {

		if _, ok := hostToName[fqdn]; ok {
			continue
		}

		weight := comparePGNamedHostInfo(existHostInfo, newHostInfo, nameToHost, hostMasterName)
		if weight == 0 {
			continue
		}

		stepNameToHost := copyMapStringString(nameToHost)
		stepHostToName := copyMapStringString(hostToName)
		stepCompareMap := copyMapIntString(compareMap)

		stepNameToHost[newHostInfo.name] = fqdn
		stepHostToName[fqdn] = newHostInfo.name
		stepCompareMap[itm] = fqdn

		compareResults = append(compareResults, generatePGNamedHostsInfoMaps(existsHostsInfo, newHostsInfo, stepNameToHost, stepHostToName, stepCompareMap, itm+1, hostMasterName)...)
	}

	return compareResults
}

func comparePGNamedHostsInfo(existsHostsInfo map[string]*pgHostInfo, nameToHost map[string]string, newHostsInfo []*pgHostInfo, hostMasterName string) (compareMap map[int]string, hostToName map[string]string) {
	compareMap = make(map[int]string)
	hostToName = make(map[string]string)

	weight := 0

	compareResults := generatePGNamedHostsInfoMaps(existsHostsInfo, newHostsInfo, nameToHost, make(map[string]string), make(map[int]string), 0, hostMasterName)

	for _, compareResult := range compareResults {
		stepWeight := comparePGNamedHostsInfoWeight(existsHostsInfo, newHostsInfo, compareResult, hostMasterName)

		if stepWeight > weight {
			weight = stepWeight
			compareMap = compareResult.compareMap
			hostToName = compareResult.hostToName
		}
	}

	return compareMap, hostToName
}

// row idx (in newHostsInfo) -> FQDN
func comparePGNoNamedHostsInfo(existingHostsInfo map[string]*pgHostInfo, newHostsInfo []*pgHostInfo) map[int]string {
	compareMap := make(map[int]string)
	visitedHostNames := make(map[string]struct{})

	for i, newHostInfo := range newHostsInfo {
		for _, existingHostInfo := range existingHostsInfo {
			if _, ok := visitedHostNames[existingHostInfo.fqdn]; ok {
				continue
			}
			if matchesPGNoNamedHostInfo(existingHostInfo, newHostInfo) {
				visitedHostNames[existingHostInfo.fqdn] = struct{}{}
				compareMap[i] = existingHostInfo.fqdn
				break
			}
		}
	}

	return compareMap
}

func loadExistingPGHostsInfo(currentHosts []*postgresql.Host, oldHosts []interface{}) (map[string]*pgHostInfo, error) {
	hostsInfo := make(map[string]*pgHostInfo)

	for i, h := range currentHosts {
		hostsInfo[h.Name] = &pgHostInfo{
			fqdn:                 h.Name,
			zone:                 h.ZoneId,
			subnetID:             h.SubnetId,
			role:                 h.Role,
			oldAssignPublicIP:    h.AssignPublicIp,
			oldPriority:          int(h.Priority.GetValue()),
			oldReplicationSource: h.ReplicationSource,

			rowNumber: i,
		}
	}

	for _, hostOldInfo := range oldHosts {
		hoi := objx.New(hostOldInfo)
		if hoi == nil {
			return nil, fmt.Errorf("PostgreSQL.host: failed to read hosts info %v", hostsInfo)
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

func validateNewPGHostsInfo(newHostsInfo []*pgHostInfo, isUpdate bool) (bool, error) {
	uniqueNames := make(map[string]struct{})
	haveHostWithName := false
	haveHostWithoutName := false

	for _, nhi := range newHostsInfo {
		name := nhi.name
		if name == "" {
			haveHostWithoutName = true
		} else {
			haveHostWithName = true
			if _, ok := uniqueNames[name]; ok && isUpdate {
				return haveHostWithName, fmt.Errorf("PostgreSQL.host: name is duplicate %v", name)
			}
			uniqueNames[name] = struct{}{}
		}
	}

	if haveHostWithName && haveHostWithoutName && isUpdate {
		return haveHostWithName, fmt.Errorf("names should be set for all hosts or unset for all host")
	}

	return haveHostWithName, nil
}

type comparePGHostsInfoResult struct {
	hostsInfo        map[string]*pgHostInfo
	createHostsInfo  []*pgHostInfo
	haveHostWithName bool
	// when hierarchyExists is true - we cannot change replication source graph in a single round
	// because we don't know all FQDNs.
	hierarchyExists bool
	hostMasterName  string
}

func comparePGHostsInfo(d *schema.ResourceData, currentHosts []*postgresql.Host, isUpdate bool) (*comparePGHostsInfoResult, error) {
	oldHosts, newHosts := d.GetChange("host")

	// actual hosts configuration (enriched with 'name', when available): fqdn -> *myHostInfo
	existingHostsInfo, err := loadExistingPGHostsInfo(currentHosts, oldHosts.([]interface{}))
	if err != nil {
		return nil, err
	}

	// expected hosts configuration: []*pgHostInfo
	newHostsInfo, err := loadNewPGHostsInfo(newHosts.([]interface{}))
	if err != nil {
		return nil, err
	}

	hostMasterName := objx.New(d.Get("host_master_name")).Value().Str()
	haveHostWithName, err := validateNewPGHostsInfo(newHostsInfo, isUpdate)
	if err != nil {
		return nil, err
	}

	nameToHost := make(map[string]string)
	for fqdn, hi := range existingHostsInfo {
		if hi.name != "" {
			nameToHost[hi.name] = fqdn
		}
	}

	createHostsInfoPrepare := make([]*pgHostInfo, 0)
	log.Printf("[DEBUG] haveHostWithName: %t", haveHostWithName)
	if haveHostWithName {
		// find best mapping from existingHostsInfo to newHostsInfo
		compareMap, hostToName := comparePGNamedHostsInfo(existingHostsInfo, nameToHost, newHostsInfo, hostMasterName)

		log.Println("[DEBUG] iterate over newHostsInfo")
		for i, newHostInfo := range newHostsInfo {
			if existHostFqdn, ok := compareMap[i]; ok {
				existHostInfo := existingHostsInfo[existHostFqdn]

				existHostInfo.name = newHostInfo.name
				existHostInfo.rowNumber = newHostInfo.rowNumber
				existHostInfo.newReplicationSourceName = newHostInfo.newReplicationSourceName
				existHostInfo.newAssignPublicIP = newHostInfo.newAssignPublicIP
				existHostInfo.newPriority = newHostInfo.newPriority
				existHostInfo.inTargetSet = true

				nameToHost[existHostInfo.name] = existHostInfo.fqdn
			} else {
				createHostsInfoPrepare = append(createHostsInfoPrepare, newHostInfo)
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

		createHostsInfo := make([]*pgHostInfo, 0)
		hierarchyExists := false
		for _, newHostInfo := range createHostsInfoPrepare {
			if newHostInfo.newReplicationSourceName == "" {
				createHostsInfo = append(createHostsInfo, newHostInfo)
			} else if fqdn, ok := nameToHost[newHostInfo.newReplicationSourceName]; ok {
				newHostInfo.newReplicationSource = fqdn
				createHostsInfo = append(createHostsInfo, newHostInfo)
			} else {
				hierarchyExists = true
			}
		}

		return &comparePGHostsInfoResult{
			hostMasterName:   hostMasterName,
			haveHostWithName: haveHostWithName,
			createHostsInfo:  createHostsInfo,
			hierarchyExists:  hierarchyExists,
			hostsInfo:        existingHostsInfo,
		}, nil
	}

	compareMap := comparePGNoNamedHostsInfo(existingHostsInfo, newHostsInfo)
	for i, newHostInfo := range newHostsInfo {
		if existingHostFqdn, ok := compareMap[i]; ok {
			log.Printf("[DEBUG] host %s exists", existingHostFqdn)
			existHostInfo := existingHostsInfo[existingHostFqdn]
			existHostInfo.rowNumber = newHostsInfo[i].rowNumber
			existHostInfo.inTargetSet = true
		} else {
			log.Printf("[DEBUG] should create host %v", newHostInfo)
			createHostsInfoPrepare = append(createHostsInfoPrepare, newHostInfo)
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
		}
	}

	return &comparePGHostsInfoResult{
		hostMasterName:   hostMasterName,
		haveHostWithName: haveHostWithName,
		createHostsInfo:  createHostsInfoPrepare,
		hostsInfo:        existingHostsInfo,
	}, nil
}

func flattenPGHostsInfo(d *schema.ResourceData, hs []*postgresql.Host) ([]*pgHostInfo, error) {
	compareHostsInfo, err := comparePGHostsInfo(d, hs, false)
	if err != nil {
		return nil, err
	}
	orderedHostsInfo := make([]*pgHostInfo, 0, len(compareHostsInfo.hostsInfo))
	for _, hostInfo := range compareHostsInfo.hostsInfo {
		orderedHostsInfo = append(orderedHostsInfo, hostInfo)
	}
	sort.Slice(orderedHostsInfo, func(i, j int) bool {
		if orderedHostsInfo[i].inTargetSet == orderedHostsInfo[j].inTargetSet {
			return orderedHostsInfo[i].rowNumber < orderedHostsInfo[j].rowNumber
		}
		return orderedHostsInfo[i].inTargetSet
	})

	return orderedHostsInfo, nil
}

func getMasterHostname(orderedHostsInfo []*pgHostInfo) string {
	for _, hostInfo := range orderedHostsInfo {
		if hostInfo.name != "" && hostInfo.role == postgresql.Host_MASTER {
			return hostInfo.name
		}
	}
	return ""
}

func flattenPGHostsFromHostInfos(orderedHostsInfo []*pgHostInfo, isDataSource bool) []map[string]interface{} {
	hosts := []map[string]interface{}{}
	for _, hostInfo := range orderedHostsInfo {
		m := map[string]interface{}{}

		m["zone"] = hostInfo.zone
		m["subnet_id"] = hostInfo.subnetID
		m["assign_public_ip"] = hostInfo.oldAssignPublicIP
		m["fqdn"] = hostInfo.fqdn
		m["role"] = hostInfo.role.String()
		m["priority"] = hostInfo.oldPriority
		m["replication_source"] = hostInfo.oldReplicationSource
		if !isDataSource {
			m["name"] = hostInfo.name
			m["replication_source_name"] = hostInfo.oldReplicationSourceName
		}
		hosts = append(hosts, m)
	}

	return hosts
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

func expandPGConfigSpec(d *schema.ResourceData) (cs *postgresql.ConfigSpec, updateFieldConfigName string, err error) {
	cs = &postgresql.ConfigSpec{}

	if v, ok := d.GetOk("config.0.version"); ok {
		cs.Version = v.(string)
	}

	if v, ok := d.GetOkExists("config.0.autofailover"); ok {
		cs.Autofailover = &wrappers.BoolValue{Value: v.(bool)}
	}

	if v, ok := d.GetOkExists("config.0.backup_retain_period_days"); ok {
		cs.BackupRetainPeriodDays = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	poolerConfig, err := expandPGPoolerConfig(d)
	if err != nil {
		return nil, updateFieldConfigName, err
	}
	cs.PoolerConfig = poolerConfig

	resources, err := expandPGResources(d)
	if err != nil {
		return nil, updateFieldConfigName, err
	}
	cs.Resources = resources

	cs.BackupWindowStart = expandPGBackupWindowStart(d)
	cs.Access = expandPGAccess(d)
	cs.PerformanceDiagnostics = expandPGPerformanceDiagnostics(d)

	updateFieldConfigName, err = expandPGConfigSpecSettings(d, cs)
	if err != nil {
		return nil, updateFieldConfigName, err
	}

	return cs, updateFieldConfigName, nil
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

	cnt := d.Get("user.#").(int)
	for i := 0; i < cnt; i++ {
		user, err := expandPGUserNew(d, fmt.Sprintf("user.%v.", i))
		if err != nil {
			return nil, err
		}

		out = append(out, user)
	}

	return out, nil
}

// pgUserForCreate get users for create
func pgUserForCreate(d *schema.ResourceData, currUsers []*postgresql.User) (usersForCreate []*postgresql.UserSpec, err error) {
	currentUser := make(map[string]struct{})
	for _, v := range currUsers {
		currentUser[v.Name] = struct{}{}
	}
	usersForCreate = make([]*postgresql.UserSpec, 0)

	cnt := d.Get("user.#").(int)
	for i := 0; i < cnt; i++ {
		_, ok := currentUser[d.Get(fmt.Sprintf("user.%v.name", i)).(string)]
		if !ok {
			user, err := expandPGUserNew(d, fmt.Sprintf("user.%v.", i))
			if err != nil {
				return nil, err
			}
			user.Grants = make([]string, 0)
			user.Permissions = make([]*postgresql.Permission, 0)
			usersForCreate = append(usersForCreate, user)
		}
	}

	return usersForCreate, nil
}

// expandPGUserNew expand to new object from schema.ResourceData
// path like "user.3."
func expandPGUserNew(d *schema.ResourceData, path string) (*postgresql.UserSpec, error) {
	return expandPGUser(d, &postgresql.UserSpec{}, path)
}

// expandPGUser expand to exists object from schema.ResourceData
// path like "user.3."
func expandPGUser(d *schema.ResourceData, user *postgresql.UserSpec, path string) (*postgresql.UserSpec, error) {

	if v, ok := d.GetOkExists(path + "name"); ok {
		user.Name = v.(string)
	}

	if v, ok := d.GetOkExists(path + "password"); ok {
		user.Password = v.(string)
	}

	if v, ok := d.GetOkExists(path + "login"); ok {
		user.Login = &wrappers.BoolValue{Value: v.(bool)}
	}

	if v, ok := d.GetOkExists(path + "conn_limit"); ok {
		user.ConnLimit = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	if v, ok := d.GetOkExists(path + "permission"); ok {
		permissions, err := expandPGUserPermissions(v.(*schema.Set))
		if err != nil {
			return nil, err
		}
		user.Permissions = permissions
	}

	if v, ok := d.GetOkExists(path + "grants"); ok {
		gs, err := expandPGUserGrants(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		user.Grants = gs
	}

	if _, ok := d.GetOkExists(path + "settings"); ok {
		if user.Settings == nil {
			user.Settings = &postgresql.UserSettings{}
		}

		err := expandResourceGenerate(mdbPGUserSettingsFieldsInfo, d, user.Settings, path+"settings.", true)
		if err != nil {
			return nil, err
		}

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
	host := &PostgreSQLHostSpec{HostSpec: hostSpec}
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
		host.Fqdn = v.(string)
	}

	if v, ok := m["replication_source_name"]; ok {
		host.HostSpec.ReplicationSource = v.(string)
	}
	if v, ok := m["priority"]; ok {
		host.HostSpec.Priority = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return host, nil
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
		out.Extensions = expandPGExtensions(es)
	}

	return out, nil
}

func expandPGExtensions(es []interface{}) []*postgresql.Extension {
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

	return out
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

func expandPGPerformanceDiagnostics(d *schema.ResourceData) *postgresql.PerformanceDiagnostics {

	if _, ok := d.GetOkExists("config.0.performance_diagnostics"); !ok {
		return nil
	}

	out := &postgresql.PerformanceDiagnostics{}

	if v, ok := d.GetOk("config.0.performance_diagnostics.0.enabled"); ok {
		out.Enabled = v.(bool)
	}

	if v, ok := d.GetOk("config.0.performance_diagnostics.0.sessions_sampling_interval"); ok {
		out.SessionsSamplingInterval = int64(v.(int))
	}

	if v, ok := d.GetOk("config.0.performance_diagnostics.0.statements_sampling_interval"); ok {
		out.StatementsSamplingInterval = int64(v.(int))
	}

	return out
}

func expandPGAccess(d *schema.ResourceData) *postgresql.Access {
	out := &postgresql.Access{}

	if v, ok := d.GetOk("config.0.access.0.data_lens"); ok {
		out.DataLens = v.(bool)
	}

	if v, ok := d.GetOk("config.0.access.0.web_sql"); ok {
		out.WebSql = v.(bool)
	}

	if v, ok := d.GetOk("config.0.access.0.serverless"); ok {
		out.Serverless = v.(bool)
	}

	return out
}

func flattenPGMaintenanceWindow(mw *postgresql.MaintenanceWindow) ([]interface{}, error) {
	maintenanceWindow := map[string]interface{}{}
	if mw != nil {
		switch p := mw.GetPolicy().(type) {
		case *postgresql.MaintenanceWindow_Anytime:
			maintenanceWindow["type"] = "ANYTIME"
			// do nothing
		case *postgresql.MaintenanceWindow_WeeklyMaintenanceWindow:
			maintenanceWindow["type"] = "WEEKLY"
			maintenanceWindow["hour"] = p.WeeklyMaintenanceWindow.Hour
			maintenanceWindow["day"] = postgresql.WeeklyMaintenanceWindow_WeekDay_name[int32(p.WeeklyMaintenanceWindow.GetDay())]
		default:
			return nil, fmt.Errorf("unsupported PostgreSQL maintenance policy type")
		}
	}

	return []interface{}{maintenanceWindow}, nil
}

func expandPGMaintenanceWindow(d *schema.ResourceData) (*postgresql.MaintenanceWindow, error) {
	if _, ok := d.GetOkExists("maintenance_window"); !ok {
		return nil, nil
	}

	out := &postgresql.MaintenanceWindow{}
	typeMW, _ := d.GetOk("maintenance_window.0.type")
	if typeMW == "ANYTIME" {
		if hour, ok := d.GetOk("maintenance_window.0.hour"); ok && hour != "" {
			return nil, fmt.Errorf("hour should be not set, when using ANYTIME")
		}
		if day, ok := d.GetOk("maintenance_window.0.day"); ok && day != "" {
			return nil, fmt.Errorf("day should be not set, when using ANYTIME")
		}
		out.Policy = &postgresql.MaintenanceWindow_Anytime{
			Anytime: &postgresql.AnytimeMaintenanceWindow{},
		}
	} else if typeMW == "WEEKLY" {
		hour := d.Get("maintenance_window.0.hour").(int)
		dayString := d.Get("maintenance_window.0.day").(string)

		day, ok := postgresql.WeeklyMaintenanceWindow_WeekDay_value[dayString]
		if !ok || day == 0 {
			return nil, fmt.Errorf(`day value should be one of ("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN")`)
		}

		out.Policy = &postgresql.MaintenanceWindow_WeeklyMaintenanceWindow{
			WeeklyMaintenanceWindow: &postgresql.WeeklyMaintenanceWindow{
				Hour: int64(hour),
				Day:  postgresql.WeeklyMaintenanceWindow_WeekDay(day),
			},
		}
	} else {
		return nil, fmt.Errorf("maintenance_window.0.type should be ANYTIME or WEEKLY")
	}

	return out, nil
}

func pgMaintenanceWindowSchemaValidateFunc(v interface{}, k string) (s []string, es []error) {
	dayString := v.(string)
	day, ok := postgresql.WeeklyMaintenanceWindow_WeekDay_value[dayString]
	if !ok || day == 0 {
		es = append(es, fmt.Errorf(`expected %s value should be one of ("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"). Current value is %v`, k, v))
		return
	}

	return
}

func expandPGConfigSpecSettings(d *schema.ResourceData, configSpec *postgresql.ConfigSpec) (updateFieldConfigName string, err error) {
	version := configSpec.Version

	path := "config.0.postgresql_config"

	if _, ok := d.GetOkExists(path); ok {

		var sharedPreloadLibraries []int32
		sharedPreloadLibValue, ok := d.GetOkExists(path + ".shared_preload_libraries")
		if ok {
			splValue := sharedPreloadLibValue.(string)

			for _, sv := range strings.Split(splValue, ",") {

				i, err := mdbPGSettingsFieldsInfo.stringToInt("shared_preload_libraries", sv)
				if err != nil {
					return updateFieldConfigName, err
				}
				if i != nil {
					sharedPreloadLibraries = append(sharedPreloadLibraries, int32(*i))
				}
			}
		}

		// TODO refactor it
		var pgConfig interface{}
		if version == "10" {
			cfg := &postgresql.ConfigSpec_PostgresqlConfig_10{
				PostgresqlConfig_10: &config.PostgresqlConfig10{},
			}
			if len(sharedPreloadLibraries) > 0 {
				for _, v := range sharedPreloadLibraries {
					cfg.PostgresqlConfig_10.SharedPreloadLibraries = append(cfg.PostgresqlConfig_10.SharedPreloadLibraries, config.PostgresqlConfig10_SharedPreloadLibraries(v))
				}
			}
			pgConfig = cfg.PostgresqlConfig_10
			configSpec.PostgresqlConfig = cfg
			updateFieldConfigName = "postgresql_config_10"
		} else if version == "10-1c" {
			cfg := &postgresql.ConfigSpec_PostgresqlConfig_10_1C{
				PostgresqlConfig_10_1C: &config.PostgresqlConfig10_1C{},
			}
			if len(sharedPreloadLibraries) > 0 {
				for _, v := range sharedPreloadLibraries {
					cfg.PostgresqlConfig_10_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_10_1C.SharedPreloadLibraries, config.PostgresqlConfig10_1C_SharedPreloadLibraries(v))
				}
			}
			pgConfig = cfg.PostgresqlConfig_10_1C
			configSpec.PostgresqlConfig = cfg
			updateFieldConfigName = "postgresql_config_10_1c"
		} else if version == "11" {
			cfg := &postgresql.ConfigSpec_PostgresqlConfig_11{
				PostgresqlConfig_11: &config.PostgresqlConfig11{},
			}
			if len(sharedPreloadLibraries) > 0 {
				for _, v := range sharedPreloadLibraries {
					cfg.PostgresqlConfig_11.SharedPreloadLibraries = append(cfg.PostgresqlConfig_11.SharedPreloadLibraries, config.PostgresqlConfig11_SharedPreloadLibraries(v))
				}
			}
			pgConfig = cfg.PostgresqlConfig_11
			configSpec.PostgresqlConfig = cfg
			updateFieldConfigName = "postgresql_config_11"
		} else if version == "11-1c" {
			cfg := &postgresql.ConfigSpec_PostgresqlConfig_11_1C{
				PostgresqlConfig_11_1C: &config.PostgresqlConfig11_1C{},
			}
			if len(sharedPreloadLibraries) > 0 {
				for _, v := range sharedPreloadLibraries {
					cfg.PostgresqlConfig_11_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_11_1C.SharedPreloadLibraries, config.PostgresqlConfig11_1C_SharedPreloadLibraries(v))
				}
			}
			pgConfig = cfg.PostgresqlConfig_11_1C
			configSpec.PostgresqlConfig = cfg
			updateFieldConfigName = "postgresql_config_11_1c"
		} else if version == "12-1c" {
			cfg := &postgresql.ConfigSpec_PostgresqlConfig_12_1C{
				PostgresqlConfig_12_1C: &config.PostgresqlConfig12_1C{},
			}
			if len(sharedPreloadLibraries) > 0 {
				for _, v := range sharedPreloadLibraries {
					cfg.PostgresqlConfig_12_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_12_1C.SharedPreloadLibraries, config.PostgresqlConfig12_1C_SharedPreloadLibraries(v))
				}
			}
			pgConfig = cfg.PostgresqlConfig_12_1C
			configSpec.PostgresqlConfig = cfg
			updateFieldConfigName = "postgresql_config_12_1c"
		} else if version == "12" {
			cfg := &postgresql.ConfigSpec_PostgresqlConfig_12{
				PostgresqlConfig_12: &config.PostgresqlConfig12{},
			}
			if len(sharedPreloadLibraries) > 0 {
				for _, v := range sharedPreloadLibraries {
					cfg.PostgresqlConfig_12.SharedPreloadLibraries = append(cfg.PostgresqlConfig_12.SharedPreloadLibraries, config.PostgresqlConfig12_SharedPreloadLibraries(v))
				}
			}
			pgConfig = cfg.PostgresqlConfig_12
			configSpec.PostgresqlConfig = cfg
			updateFieldConfigName = "postgresql_config_12"
		} else if version == "13" {
			cfg := &postgresql.ConfigSpec_PostgresqlConfig_13{
				PostgresqlConfig_13: &config.PostgresqlConfig13{},
			}
			if len(sharedPreloadLibraries) > 0 {
				for _, v := range sharedPreloadLibraries {
					cfg.PostgresqlConfig_13.SharedPreloadLibraries = append(cfg.PostgresqlConfig_13.SharedPreloadLibraries, config.PostgresqlConfig13_SharedPreloadLibraries(v))
				}
			}
			pgConfig = cfg.PostgresqlConfig_13
			configSpec.PostgresqlConfig = cfg
			updateFieldConfigName = "postgresql_config_13"
		} else if version == "14" {
			cfg := &postgresql.ConfigSpec_PostgresqlConfig_14{
				PostgresqlConfig_14: &config.PostgresqlConfig14{},
			}
			if len(sharedPreloadLibraries) > 0 {
				for _, v := range sharedPreloadLibraries {
					cfg.PostgresqlConfig_14.SharedPreloadLibraries = append(cfg.PostgresqlConfig_14.SharedPreloadLibraries, config.PostgresqlConfig14_SharedPreloadLibraries(v))
				}
			}
			pgConfig = cfg.PostgresqlConfig_14
			configSpec.PostgresqlConfig = cfg
			updateFieldConfigName = "postgresql_config_14"
		} else {
			return updateFieldConfigName, nil
		}

		err := expandResourceGenerate(mdbPGSettingsFieldsInfo, d, pgConfig, path+".", true)

		if err != nil {
			return updateFieldConfigName, err
		}

	}

	return updateFieldConfigName, nil
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

var mdbPGUserSettingsTransactionIsolationName = map[int]string{
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_UNSPECIFIED):      "unspecified",
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_READ_UNCOMMITTED): "read uncommitted",
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_READ_COMMITTED):   "read committed",
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_REPEATABLE_READ):  "repeatable read",
	int(postgresql.UserSettings_TRANSACTION_ISOLATION_SERIALIZABLE):     "serializable",
}
var mdbPGUserSettingsSynchronousCommitName = map[int]string{
	int(postgresql.UserSettings_SYNCHRONOUS_COMMIT_UNSPECIFIED):  "unspecified",
	int(postgresql.UserSettings_SYNCHRONOUS_COMMIT_ON):           "on",
	int(postgresql.UserSettings_SYNCHRONOUS_COMMIT_OFF):          "off",
	int(postgresql.UserSettings_SYNCHRONOUS_COMMIT_LOCAL):        "local",
	int(postgresql.UserSettings_SYNCHRONOUS_COMMIT_REMOTE_WRITE): "remote write",
	int(postgresql.UserSettings_SYNCHRONOUS_COMMIT_REMOTE_APPLY): "remote apply",
}
var mdbPGUserSettingsLogStatementName = map[int]string{
	int(postgresql.UserSettings_LOG_STATEMENT_UNSPECIFIED): "unspecified",
	int(postgresql.UserSettings_LOG_STATEMENT_NONE):        "none",
	int(postgresql.UserSettings_LOG_STATEMENT_DDL):         "ddl",
	int(postgresql.UserSettings_LOG_STATEMENT_MOD):         "mod",
	int(postgresql.UserSettings_LOG_STATEMENT_ALL):         "all",
}

var mdbPGUserSettingsFieldsInfo = newObjectFieldsInfo().
	addType(postgresql.UserSettings{}).
	addIDefault("log_min_duration_statement", -1).
	addEnumHumanNames("default_transaction_isolation", mdbPGUserSettingsTransactionIsolationName,
		postgresql.UserSettings_TransactionIsolation_name).
	addEnumHumanNames("synchronous_commit", mdbPGUserSettingsSynchronousCommitName,
		postgresql.UserSettings_SynchronousCommit_name).
	addEnumHumanNames("log_statement", mdbPGUserSettingsLogStatementName,
		postgresql.UserSettings_LogStatement_name)

var mdbPGSettingsFieldsInfo = newObjectFieldsInfo().
	addType(config.PostgresqlConfig14{}).
	addType(config.PostgresqlConfig13{}).
	addType(config.PostgresqlConfig12{}).
	addType(config.PostgresqlConfig12_1C{}).
	addType(config.PostgresqlConfig11{}).
	addType(config.PostgresqlConfig11_1C{}).
	addType(config.PostgresqlConfig10{}).
	addType(config.PostgresqlConfig10_1C{}).
	addEnumGeneratedNames("wal_level", config.PostgresqlConfig13_WalLevel_name).
	addEnumGeneratedNames("synchronous_commit", config.PostgresqlConfig13_SynchronousCommit_name).
	addEnumGeneratedNames("constraint_exclusion", config.PostgresqlConfig13_ConstraintExclusion_name).
	addEnumGeneratedNames("force_parallel_mode", config.PostgresqlConfig13_ForceParallelMode_name).
	addEnumGeneratedNames("client_min_messages", config.PostgresqlConfig13_LogLevel_name).
	addEnumGeneratedNames("log_min_messages", config.PostgresqlConfig13_LogLevel_name).
	addEnumGeneratedNames("log_min_error_statement", config.PostgresqlConfig13_LogLevel_name).
	addEnumGeneratedNames("log_error_verbosity", config.PostgresqlConfig13_LogErrorVerbosity_name).
	addEnumGeneratedNames("log_statement", config.PostgresqlConfig13_LogStatement_name).
	addEnumGeneratedNames("default_transaction_isolation", config.PostgresqlConfig13_TransactionIsolation_name).
	addEnumGeneratedNames("bytea_output", config.PostgresqlConfig13_ByteaOutput_name).
	addEnumGeneratedNames("xmlbinary", config.PostgresqlConfig13_XmlBinary_name).
	addEnumGeneratedNames("xmloption", config.PostgresqlConfig13_XmlOption_name).
	addEnumGeneratedNames("backslash_quote", config.PostgresqlConfig13_BackslashQuote_name).
	addEnumGeneratedNames("plan_cache_mode", config.PostgresqlConfig13_PlanCacheMode_name).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig13_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare).
	addEnumGeneratedNames("pg_hint_plan_debug_print", config.PostgresqlConfig13_PgHintPlanDebugPrint_name).
	addEnumGeneratedNames("pg_hint_plan_message_level", config.PostgresqlConfig13_LogLevel_name).
	addEnumGeneratedNames("wal_level", config.PostgresqlConfig14_WalLevel_name).
	addEnumGeneratedNames("synchronous_commit", config.PostgresqlConfig14_SynchronousCommit_name).
	addEnumGeneratedNames("constraint_exclusion", config.PostgresqlConfig14_ConstraintExclusion_name).
	addEnumGeneratedNames("force_parallel_mode", config.PostgresqlConfig14_ForceParallelMode_name).
	addEnumGeneratedNames("client_min_messages", config.PostgresqlConfig14_LogLevel_name).
	addEnumGeneratedNames("log_min_messages", config.PostgresqlConfig14_LogLevel_name).
	addEnumGeneratedNames("log_min_error_statement", config.PostgresqlConfig14_LogLevel_name).
	addEnumGeneratedNames("log_error_verbosity", config.PostgresqlConfig14_LogErrorVerbosity_name).
	addEnumGeneratedNames("log_statement", config.PostgresqlConfig14_LogStatement_name).
	addEnumGeneratedNames("default_transaction_isolation", config.PostgresqlConfig14_TransactionIsolation_name).
	addEnumGeneratedNames("bytea_output", config.PostgresqlConfig14_ByteaOutput_name).
	addEnumGeneratedNames("xmlbinary", config.PostgresqlConfig14_XmlBinary_name).
	addEnumGeneratedNames("xmloption", config.PostgresqlConfig14_XmlOption_name).
	addEnumGeneratedNames("backslash_quote", config.PostgresqlConfig14_BackslashQuote_name).
	addEnumGeneratedNames("plan_cache_mode", config.PostgresqlConfig14_PlanCacheMode_name).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig14_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare).
	addEnumGeneratedNames("pg_hint_plan_debug_print", config.PostgresqlConfig14_PgHintPlanDebugPrint_name).
	addEnumGeneratedNames("pg_hint_plan_message_level", config.PostgresqlConfig14_LogLevel_name)
