package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/objx"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
)

type PostgreSQLHostSpec struct {
	HostSpec *postgresql.HostSpec
	Fqdn     string
}

type PGAuditSettings struct {
	Log []string `json:"log"`
}

func flattenPGClusterConfig(c *postgresql.ClusterConfig) ([]interface{}, error) {
	settings, err := flattenPGSettings(c)
	if err != nil {
		return nil, err
	}

	out := map[string]interface{}{}
	out["autofailover"] = c.GetAutofailover().GetValue()
	out["version"] = c.Version
	out["pooler_config"] = flattenPGPoolerConfig(c.PoolerConfig)
	out["resources"] = flattenPGResources(c.Resources)
	out["backup_window_start"] = flattenMDBBackupWindowStart(c.BackupWindowStart)
	out["backup_retain_period_days"] = c.BackupRetainPeriodDays.GetValue()
	out["performance_diagnostics"] = flattenPGPerformanceDiagnostics(c.PerformanceDiagnostics)
	out["disk_size_autoscaling"] = flattenPGDiskSizeAutoscaling(c.DiskSizeAutoscaling)
	out["access"] = flattenPGAccess(c.Access)
	out["postgresql_config"] = settings

	return []interface{}{out}, nil
}

func flattenPGPoolerConfig(c *postgresql.ConnectionPoolerConfig) []interface{} {
	if c == nil {
		return nil
	}

	out := map[string]interface{}{}
	out["pool_discard"] = c.GetPoolDiscard().GetValue()
	out["pooling_mode"] = c.GetPoolingMode().String()

	return []interface{}{out}
}

func flattenPGResources(r *postgresql.Resources) []interface{} {
	out := map[string]interface{}{}
	out["resource_preset_id"] = r.ResourcePresetId
	out["disk_size"] = toGigabytes(r.DiskSize)
	out["disk_type_id"] = r.DiskTypeId

	return []interface{}{out}
}

func flattenPGPerformanceDiagnostics(p *postgresql.PerformanceDiagnostics) []interface{} {
	if p == nil {
		return nil
	}

	out := map[string]interface{}{}

	out["enabled"] = p.Enabled
	out["sessions_sampling_interval"] = int(p.SessionsSamplingInterval)
	out["statements_sampling_interval"] = int(p.StatementsSamplingInterval)

	return []interface{}{out}
}

func flattenPGDiskSizeAutoscaling(p *postgresql.DiskSizeAutoscaling) []interface{} {
	if p == nil {
		return nil
	}

	out := map[string]interface{}{}

	out["disk_size_limit"] = toGigabytes(p.DiskSizeLimit)
	out["planned_usage_threshold"] = int(p.PlannedUsageThreshold)
	out["emergency_usage_threshold"] = int(p.EmergencyUsageThreshold)

	return []interface{}{out}
}

func flattenPGSettingsSPL(settings map[string]string, fieldsInfo *objectFieldsInfo, c *postgresql.ClusterConfig) map[string]string {
	splEnums := convertPGSPLtoInts(c)
	spl, _ := fieldsInfo.intSliceToString("shared_preload_libraries", splEnums)
	if settings == nil {
		settings = make(map[string]string)
	}
	settings["shared_preload_libraries"] = spl
	return settings
}

func convertPGSPLtoInts(c *postgresql.ClusterConfig) []int32 {
	out := []int32{}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_18); ok {
		for _, v := range cf.PostgresqlConfig_18.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_18_1C); ok {
		for _, v := range cf.PostgresqlConfig_18_1C.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_17); ok {
		for _, v := range cf.PostgresqlConfig_17.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_17_1C); ok {
		for _, v := range cf.PostgresqlConfig_17_1C.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_16); ok {
		for _, v := range cf.PostgresqlConfig_16.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_16_1C); ok {
		for _, v := range cf.PostgresqlConfig_16_1C.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_15); ok {
		for _, v := range cf.PostgresqlConfig_15.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_15_1C); ok {
		for _, v := range cf.PostgresqlConfig_15_1C.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_14); ok {
		for _, v := range cf.PostgresqlConfig_14.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_14_1C); ok {
		for _, v := range cf.PostgresqlConfig_14_1C.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_13); ok {
		for _, v := range cf.PostgresqlConfig_13.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_13_1C); ok {
		for _, v := range cf.PostgresqlConfig_13_1C.UserConfig.SharedPreloadLibraries {
			out = append(out, int32(v))
		}
	}
	return out
}

func flattenPGSettings(c *postgresql.ClusterConfig) (map[string]string, error) {
	// TODO refactor it using generics
	settingsFieldsInfo, err := getMdbPGSettingsFieldsInfo(c.Version)
	if err != nil {
		return nil, err
	}

	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_18); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_18.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_18_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_18_1C.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_17); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_17.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_17_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_17_1C.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_16); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_16.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_16_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_16_1C.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_15); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_15.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_15_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_15_1C.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_14); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_14.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_14_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_14_1C.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_13); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_13.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}
	if cf, ok := c.PostgresqlConfig.(*postgresql.ClusterConfig_PostgresqlConfig_13_1C); ok {
		settings, err := flattenResourceGenerateMapS(cf.PostgresqlConfig_13_1C.UserConfig, false, settingsFieldsInfo, false, true, nil)
		if err != nil {
			return nil, err
		}
		settings = flattenPGSettingsSPL(settings, settingsFieldsInfo, c)
		return settings, nil
	}

	return nil, nil
}

func flattenPGAccess(a *postgresql.Access) []interface{} {
	if a == nil {
		return nil
	}

	out := map[string]interface{}{}
	out["data_lens"] = a.DataLens
	out["web_sql"] = a.WebSql
	out["serverless"] = a.Serverless
	out["data_transfer"] = a.DataTransfer

	return []interface{}{out}
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

func flattenPGUserConnectionManager(cm *postgresql.ConnectionManager) map[string]string {
	if cm == nil {
		return nil
	}
	return map[string]string{"connection_id": cm.ConnectionId}
}

type pgHostInfo struct {
	name string
	fqdn string

	zone     string
	subnetID string

	role postgresql.Host_Role

	oldAssignPublicIP        bool
	oldReplicationSource     string
	oldReplicationSourceName string

	newAssignPublicIP        bool
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
			rowNumber:                i,
			inTargetSet:              true,
		})

	}

	return hostsInfo, nil
}

func comparePGNamedHostInfo(existsHostInfo *pgHostInfo, newHostInfo *pgHostInfo, currentNameToHost map[string]string) int {
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
		if fqdn, ok := currentNameToHost[newHostInfo.newReplicationSourceName]; ok && existsHostInfo.oldReplicationSource == fqdn {
			compareWeight += 4
		}
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

func comparePGNamedHostsInfoWeight(existsHostsInfo map[string]*pgHostInfo, newHostsInfo []*pgHostInfo, compareResult pgCompareHostNameResult) int {
	weight := 0

	for i, fqdn := range compareResult.compareMap {
		weightStep := comparePGNamedHostInfo(existsHostsInfo[fqdn], newHostsInfo[i], compareResult.nameToHost)
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

func generatePGNamedHostsInfoMaps(existsHostsInfo map[string]*pgHostInfo, newHostsInfo []*pgHostInfo, nameToHost map[string]string, hostToName map[string]string, compareMap map[int]string, itm int) (compareResults []pgCompareHostNameResult) {
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
		return generatePGNamedHostsInfoMaps(existsHostsInfo, newHostsInfo, nameToHost, hostToName, compareMap, itm+1)
	}

	compareResults = append(compareResults, generatePGNamedHostsInfoMaps(existsHostsInfo, newHostsInfo, copyMapStringString(nameToHost), copyMapStringString(hostToName), copyMapIntString(compareMap), itm+1)...)

	for fqdn, existHostInfo := range existsHostsInfo {

		if _, ok := hostToName[fqdn]; ok {
			continue
		}

		weight := comparePGNamedHostInfo(existHostInfo, newHostInfo, nameToHost)
		if weight == 0 {
			continue
		}

		stepNameToHost := copyMapStringString(nameToHost)
		stepHostToName := copyMapStringString(hostToName)
		stepCompareMap := copyMapIntString(compareMap)

		stepNameToHost[newHostInfo.name] = fqdn
		stepHostToName[fqdn] = newHostInfo.name
		stepCompareMap[itm] = fqdn

		compareResults = append(compareResults, generatePGNamedHostsInfoMaps(existsHostsInfo, newHostsInfo, stepNameToHost, stepHostToName, stepCompareMap, itm+1)...)
	}

	return compareResults
}

func comparePGNamedHostsInfo(existsHostsInfo map[string]*pgHostInfo, nameToHost map[string]string, newHostsInfo []*pgHostInfo) (compareMap map[int]string, hostToName map[string]string) {
	compareMap = make(map[int]string)
	hostToName = make(map[string]string)

	weight := 0

	compareResults := generatePGNamedHostsInfoMaps(existsHostsInfo, newHostsInfo, nameToHost, make(map[string]string), make(map[int]string), 0)

	for _, compareResult := range compareResults {
		stepWeight := comparePGNamedHostsInfoWeight(existsHostsInfo, newHostsInfo, compareResult)

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
		compareMap, hostToName := comparePGNamedHostsInfo(existingHostsInfo, nameToHost, newHostsInfo)

		log.Println("[DEBUG] iterate over newHostsInfo")
		for i, newHostInfo := range newHostsInfo {
			if existHostFqdn, ok := compareMap[i]; ok {
				existHostInfo := existingHostsInfo[existHostFqdn]

				existHostInfo.name = newHostInfo.name
				existHostInfo.rowNumber = newHostInfo.rowNumber
				existHostInfo.newReplicationSourceName = newHostInfo.newReplicationSourceName
				existHostInfo.newAssignPublicIP = newHostInfo.newAssignPublicIP
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

			existHostInfo.rowNumber = newHostInfo.rowNumber
			existHostInfo.newReplicationSourceName = newHostInfo.newReplicationSourceName
			existHostInfo.newAssignPublicIP = newHostInfo.newAssignPublicIP
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

func getPostgreSQLConfigFieldName(version string) (string, error) {
	switch version {
	case "13":
		return "postgresql_config_13", nil
	case "13-1c":
		return "postgresql_config_13_1c", nil
	case "14":
		return "postgresql_config_14", nil
	case "14-1c":
		return "postgresql_config_14_1c", nil
	case "15":
		return "postgresql_config_15", nil
	case "15-1c":
		return "postgresql_config_15_1c", nil
	case "16":
		return "postgresql_config_16", nil
	case "16-1c":
		return "postgresql_config_16_1c", nil
	case "17":
		return "postgresql_config_17", nil
	case "17-1c":
		return "postgresql_config_17_1c", nil
	case "18":
		return "postgresql_config_18", nil
	case "18-1c":
		return "postgresql_config_18_1c", nil
	default:
		return "", fmt.Errorf("Unsupported postgresql version: %s", version)
	}
}

func flattenPGHostsFromHostInfos(d *schema.ResourceData, orderedHostsInfo []*pgHostInfo, isDataSource bool) []map[string]interface{} {
	isNameFieldUsed := checkNameFieldUsage(d)
	log.Printf("[DEBUG] isNameFieldUsed = %t", isNameFieldUsed)
	hosts := []map[string]interface{}{}
	for _, hostInfo := range orderedHostsInfo {
		m := map[string]interface{}{}

		m["zone"] = hostInfo.zone
		m["subnet_id"] = hostInfo.subnetID
		m["assign_public_ip"] = hostInfo.oldAssignPublicIP
		m["fqdn"] = hostInfo.fqdn
		m["role"] = hostInfo.role.String()
		m["replication_source"] = hostInfo.oldReplicationSource
		if !isDataSource && isNameFieldUsed {
			m["name"] = hostInfo.name
			m["replication_source_name"] = hostInfo.oldReplicationSourceName
		}
		hosts = append(hosts, m)
	}

	return hosts
}

func checkNameFieldUsage(d *schema.ResourceData) bool {
	log.Print("[DEBUG] checkNameFieldUsage")
	hosts := d.Get("host").([]interface{})
	if len(hosts) == 0 {
		return false
	}
	host := hosts[0].(map[string]interface{})
	log.Printf("[DEBUG] host name is '%s'", host["name"])
	return host["name"] != ""
}

func flattenPGDatabases(dbs []*postgresql.Database) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)

	for _, d := range dbs {
		m := make(map[string]interface{})
		m["name"] = d.Name
		m["owner"] = d.Owner
		m["lc_collate"] = d.LcCollate
		m["template_db"] = d.TemplateDb
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

	return hashcode.String(buf.String())
}

func expandPGParamsUpdatePath(d *schema.ResourceData, settingNames []string) ([]string, error) {
	log.Println("[DEBUG] expandPGParamsUpdatePath")
	version := d.Get("config.0.version").(string)
	pgFieldName, err := getPostgreSQLConfigFieldName(version)
	if err != nil {
		return []string{}, err
	}
	log.Print("[DEBUG] pgFieldName")

	mdbPGUpdateFieldsMap := map[string]string{
		"name":                                       "name",
		"description":                                "description",
		"labels":                                     "labels",
		"network_id":                                 "network_id",
		"config.0.version":                           "config_spec.version",
		"config.0.autofailover":                      "config_spec.autofailover",
		"config.0.pooler_config.0.pooling_mode":      "config_spec.pooler_config.pooling_mode",
		"config.0.pooler_config.0.pool_discard":      "config_spec.pooler_config.pool_discard",
		"config.0.access.0.data_lens":                "config_spec.access.data_lens",
		"config.0.access.0.web_sql":                  "config_spec.access.web_sql",
		"config.0.access.0.serverless":               "config_spec.access.serverless",
		"config.0.access.0.data_transfer":            "config_spec.access.data_transfer",
		"config.0.performance_diagnostics.0.enabled": "config_spec.performance_diagnostics.enabled",
		"config.0.performance_diagnostics.0.sessions_sampling_interval":   "config_spec.performance_diagnostics.sessions_sampling_interval",
		"config.0.performance_diagnostics.0.statements_sampling_interval": "config_spec.performance_diagnostics.statements_sampling_interval",
		"config.0.disk_size_autoscaling":                                  "config_spec.disk_size_autoscaling",
		"config.0.backup_window_start":                                    "config_spec.backup_window_start",
		"config.0.resources":                                              "config_spec.resources",
		"config.0.backup_retain_period_days":                              "config_spec.backup_retain_period_days",
		"security_group_ids":                                              "security_group_ids",
		"maintenance_window":                                              "maintenance_window",
		"deletion_protection":                                             "deletion_protection",
		"config.0.postgresql_config.shared_preload_libraries":             fmt.Sprintf("config_spec.%s.shared_preload_libraries", pgFieldName),
	}

	updatePath := []string{}
	for field, path := range mdbPGUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	for _, setting := range settingNames {
		field := fmt.Sprintf("config.0.postgresql_config.%s", setting)
		log.Printf("[DEBUG] HasChange(%s): %t", field, d.HasChange(field))
		if d.HasChange(field) {
			path := fmt.Sprintf("config_spec.%s.%s", pgFieldName, setting)
			updatePath = append(updatePath, path)
		}
	}

	return updatePath, nil
}

func expandPGConfigSpec(d *schema.ResourceData) (*postgresql.ConfigSpec, []string, error) {
	poolerConfig, err := expandPGPoolerConfig(d)
	if err != nil {
		return nil, nil, err
	}

	resources, err := expandPGResources(d)
	if err != nil {
		return nil, nil, err
	}

	cs := &postgresql.ConfigSpec{
		Version:                d.Get("config.0.version").(string),
		Autofailover:           expandPGConfigAutofailover(d),
		BackupRetainPeriodDays: expandPGBackupRetainPeriodDays(d),
		PoolerConfig:           poolerConfig,
		Resources:              resources,
		BackupWindowStart:      expandMDBBackupWindowStart(d, "config.0.backup_window_start.0"),
		Access:                 expandPGAccess(d),
		PerformanceDiagnostics: expandPGPerformanceDiagnostics(d),
		DiskSizeAutoscaling:    expandPGDiskSizeAutoscaling(d),
	}

	settingNames, err := expandPGConfigSpecSettings(d, cs)
	if err != nil {
		return nil, nil, err
	}

	return cs, settingNames, nil
}

func expandPGConfigAutofailover(d *schema.ResourceData) *wrappers.BoolValue {
	if v, ok := d.GetOkExists("config.0.autofailover"); ok {
		return &wrappers.BoolValue{Value: v.(bool)}
	}
	return nil
}

func expandPGBackupRetainPeriodDays(d *schema.ResourceData) *wrappers.Int64Value {
	if v, ok := d.GetOkExists("config.0.backup_retain_period_days"); ok {
		return &wrappers.Int64Value{Value: int64(v.(int))}
	}
	return nil
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

func expandPgAuditSettings(as string) (*postgresql.PGAuditSettings, error) {
	var auditSettings PGAuditSettings
	err := json.Unmarshal([]byte(as), &auditSettings)
	if err != nil {
		return nil, err
	}

	asl := make([]postgresql.PGAuditSettings_PGAuditSettingsLog, 0)
	for _, log := range auditSettings.Log {
		asl = append(asl, mdbPGUserSettingsPgauditName[strings.ToLower(log)])
	}

	return &postgresql.PGAuditSettings{Log: asl}, nil
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

	if v, ok := m["template_db"]; ok {
		out.TemplateDb = v.(string)
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

		out = append(out, extension)
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

func expandPGDiskSizeAutoscaling(d *schema.ResourceData) *postgresql.DiskSizeAutoscaling {

	if _, ok := d.GetOkExists("config.0.disk_size_autoscaling"); !ok {
		return nil
	}

	out := &postgresql.DiskSizeAutoscaling{}

	if v, ok := d.GetOk("config.0.disk_size_autoscaling.0.disk_size_limit"); ok {
		out.DiskSizeLimit = toBytes(v.(int))
	}

	if v, ok := d.GetOk("config.0.disk_size_autoscaling.0.planned_usage_threshold"); ok {
		out.PlannedUsageThreshold = int64(v.(int))
	}

	if v, ok := d.GetOk("config.0.disk_size_autoscaling.0.emergency_usage_threshold"); ok {
		out.EmergencyUsageThreshold = int64(v.(int))
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

	if v, ok := d.GetOk("config.0.access.0.data_transfer"); ok {
		out.DataTransfer = v.(bool)
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

func expandPGSharedPreloadLibraries(d *schema.ResourceData, version string) ([]int32, error) {
	var sharedPreloadLibraries []int32
	sharedPreloadLibValue, ok := d.GetOkExists("config.0.postgresql_config.shared_preload_libraries")
	mdbPGSettingsFieldsInfo, err := getMdbPGSettingsFieldsInfo(version)
	if err != nil {
		return []int32{}, err
	}
	if ok {
		splValue := sharedPreloadLibValue.(string)

		for _, sv := range strings.Split(splValue, ",") {

			i, err := mdbPGSettingsFieldsInfo.stringToInt("shared_preload_libraries", sv)
			if err != nil {
				return []int32{}, err
			}
			if i != nil {
				sharedPreloadLibraries = append(sharedPreloadLibraries, int32(*i))
			}
		}
	}
	return sharedPreloadLibraries, nil
}

func expandPGConfigSpecSettings(d *schema.ResourceData, configSpec *postgresql.ConfigSpec) ([]string, error) {
	if _, ok := d.GetOkExists("config.0.postgresql_config"); !ok {
		log.Println("[DEBUG] config.0.postgresql_config does not exists")
		return []string{}, nil
	}
	log.Println("[DEBUG] config.0.postgresql_config exists")
	version := configSpec.Version

	sharedPreloadLibraries, err := expandPGSharedPreloadLibraries(d, version)
	if err != nil {
		return []string{}, err
	}

	if version == "13" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_13{
			PostgresqlConfig_13: &config.PostgresqlConfig13{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_13.SharedPreloadLibraries = append(cfg.PostgresqlConfig_13.SharedPreloadLibraries, config.PostgresqlConfig13_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo13, d, cfg.PostgresqlConfig_13, "config.0.postgresql_config.", true)
	} else if version == "13-1c" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_13_1C{
			PostgresqlConfig_13_1C: &config.PostgresqlConfig13_1C{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_13_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_13_1C.SharedPreloadLibraries, config.PostgresqlConfig13_1C_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo13_1C, d, cfg.PostgresqlConfig_13_1C, "config.0.postgresql_config.", true)
	} else if version == "14" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_14{
			PostgresqlConfig_14: &config.PostgresqlConfig14{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_14.SharedPreloadLibraries = append(cfg.PostgresqlConfig_14.SharedPreloadLibraries, config.PostgresqlConfig14_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo14, d, cfg.PostgresqlConfig_14, "config.0.postgresql_config.", true)
	} else if version == "14-1c" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_14_1C{
			PostgresqlConfig_14_1C: &config.PostgresqlConfig14_1C{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_14_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_14_1C.SharedPreloadLibraries, config.PostgresqlConfig14_1C_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo14_1C, d, cfg.PostgresqlConfig_14_1C, "config.0.postgresql_config.", true)
	} else if version == "15" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_15{
			PostgresqlConfig_15: &config.PostgresqlConfig15{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_15.SharedPreloadLibraries = append(cfg.PostgresqlConfig_15.SharedPreloadLibraries, config.PostgresqlConfig15_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo15, d, cfg.PostgresqlConfig_15, "config.0.postgresql_config.", true)
	} else if version == "15-1c" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_15_1C{
			PostgresqlConfig_15_1C: &config.PostgresqlConfig15_1C{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_15_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_15_1C.SharedPreloadLibraries, config.PostgresqlConfig15_1C_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo15_1C, d, cfg.PostgresqlConfig_15_1C, "config.0.postgresql_config.", true)
	} else if version == "16" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_16{
			PostgresqlConfig_16: &config.PostgresqlConfig16{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_16.SharedPreloadLibraries = append(cfg.PostgresqlConfig_16.SharedPreloadLibraries, config.PostgresqlConfig16_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo16, d, cfg.PostgresqlConfig_16, "config.0.postgresql_config.", true)
	} else if version == "16-1c" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_16_1C{
			PostgresqlConfig_16_1C: &config.PostgresqlConfig16_1C{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_16_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_16_1C.SharedPreloadLibraries, config.PostgresqlConfig16_1C_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo16_1C, d, cfg.PostgresqlConfig_16_1C, "config.0.postgresql_config.", true)
	} else if version == "17" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_17{
			PostgresqlConfig_17: &config.PostgresqlConfig17{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_17.SharedPreloadLibraries = append(cfg.PostgresqlConfig_17.SharedPreloadLibraries, config.PostgresqlConfig17_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo17, d, cfg.PostgresqlConfig_17, "config.0.postgresql_config.", true)
	} else if version == "17-1c" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_17_1C{
			PostgresqlConfig_17_1C: &config.PostgresqlConfig17_1C{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_17_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_17_1C.SharedPreloadLibraries, config.PostgresqlConfig17_1C_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo17_1C, d, cfg.PostgresqlConfig_17_1C, "config.0.postgresql_config.", true)
	} else if version == "18" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_18{
			PostgresqlConfig_18: &config.PostgresqlConfig18{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_18.SharedPreloadLibraries = append(cfg.PostgresqlConfig_18.SharedPreloadLibraries, config.PostgresqlConfig18_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo18, d, cfg.PostgresqlConfig_18, "config.0.postgresql_config.", true)
	} else if version == "18-1c" {
		cfg := &postgresql.ConfigSpec_PostgresqlConfig_18_1C{
			PostgresqlConfig_18_1C: &config.PostgresqlConfig18_1C{},
		}
		if len(sharedPreloadLibraries) > 0 {
			for _, v := range sharedPreloadLibraries {
				cfg.PostgresqlConfig_18_1C.SharedPreloadLibraries = append(cfg.PostgresqlConfig_18_1C.SharedPreloadLibraries, config.PostgresqlConfig18_1C_SharedPreloadLibraries(v))
			}
		}
		configSpec.PostgresqlConfig = cfg
		return expandResourceGenerateNonSkippedFields(mdbPGSettingsFieldsInfo18_1C, d, cfg.PostgresqlConfig_18_1C, "config.0.postgresql_config.", true)
	}

	return []string{}, err
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

func parsePostgreSQLAuthMethod(s string) (postgresql.AuthMethod, error) {
	v, ok := postgresql.AuthMethod_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'auth_method' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(postgresql.AuthMethod_value)), s)
	}

	return postgresql.AuthMethod(v), nil
}

var mdbPGTristateBooleanName = map[string]*wrappers.BoolValue{
	"true":        wrapperspb.Bool(true),
	"false":       wrapperspb.Bool(false),
	"unspecified": nil,
}

func mdbPGResolveTristateBoolean(value *wrappers.BoolValue) string {
	if value == nil {
		return "unspecified"
	}
	if value.Value {
		return "true"
	}
	return "false"
}

func validatePasswordConfiguration(userSpec *postgresql.UserSpec) error {
	passwordSpecified := len(userSpec.Password) > 0

	isBothFieldSpecified := passwordSpecified && userSpec.GeneratePassword.GetValue()
	isAnyFieldSpecified := passwordSpecified || userSpec.GeneratePassword.GetValue()
	if userSpec.AuthMethod == postgresql.AuthMethod_AUTH_METHOD_IAM {
		if isAnyFieldSpecified {
			return fmt.Errorf("%q does not support password or generate_password", userSpec.AuthMethod.String())
		}
	}
	if isBothFieldSpecified {
		return fmt.Errorf("must specify either password or generate_password")

	}
	return nil
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
var mdbPGUserSettingsPoolModeName = map[int]string{
	int(postgresql.UserSettings_POOLING_MODE_UNSPECIFIED): "unspecified",
	int(postgresql.UserSettings_STATEMENT):                "statement",
	int(postgresql.UserSettings_TRANSACTION):              "transaction",
	int(postgresql.UserSettings_SESSION):                  "session",
}

var mdbPGUserSettingsPgauditName = map[string]postgresql.PGAuditSettings_PGAuditSettingsLog{
	"read":     postgresql.PGAuditSettings_PG_AUDIT_SETTINGS_LOG_READ,
	"write":    postgresql.PGAuditSettings_PG_AUDIT_SETTINGS_LOG_WRITE,
	"function": postgresql.PGAuditSettings_PG_AUDIT_SETTINGS_LOG_FUNCTION,
	"role":     postgresql.PGAuditSettings_PG_AUDIT_SETTINGS_LOG_ROLE,
	"ddl":      postgresql.PGAuditSettings_PG_AUDIT_SETTINGS_LOG_DDL,
	"misc":     postgresql.PGAuditSettings_PG_AUDIT_SETTINGS_LOG_MISC,
	"misc_set": postgresql.PGAuditSettings_PG_AUDIT_SETTINGS_LOG_MISC_SET,
}

var mdbPGUserSettingsFieldsInfo = newObjectFieldsInfo().
	addType(postgresql.UserSettings{}, []reflect.Type{reflect.TypeOf(&postgresql.PGAuditSettings{})}).
	addFieldInfoManually("pgaudit", true).
	addIDefault("log_min_duration_statement", -1).
	addEnumHumanNames("default_transaction_isolation", mdbPGUserSettingsTransactionIsolationName,
		postgresql.UserSettings_TransactionIsolation_name).
	addEnumHumanNames("synchronous_commit", mdbPGUserSettingsSynchronousCommitName,
		postgresql.UserSettings_SynchronousCommit_name).
	addEnumHumanNames("log_statement", mdbPGUserSettingsLogStatementName,
		postgresql.UserSettings_LogStatement_name).
	addEnumHumanNames("pool_mode", mdbPGUserSettingsPoolModeName,
		postgresql.UserSettings_PoolingMode_name)

func getMdbPGSettingsFieldsInfo(version string) (*objectFieldsInfo, error) {
	switch version {
	case "13":
		return mdbPGSettingsFieldsInfo13, nil
	case "13-1c":
		return mdbPGSettingsFieldsInfo13_1C, nil
	case "14":
		return mdbPGSettingsFieldsInfo14, nil
	case "14-1c":
		return mdbPGSettingsFieldsInfo14_1C, nil
	case "15":
		return mdbPGSettingsFieldsInfo15, nil
	case "15-1c":
		return mdbPGSettingsFieldsInfo15_1C, nil
	case "16":
		return mdbPGSettingsFieldsInfo16, nil
	case "16-1c":
		return mdbPGSettingsFieldsInfo16_1C, nil
	case "17":
		return mdbPGSettingsFieldsInfo17, nil
	case "17-1c":
		return mdbPGSettingsFieldsInfo17_1C, nil
	case "18":
		return mdbPGSettingsFieldsInfo18, nil
	case "18-1c":
		return mdbPGSettingsFieldsInfo18_1C, nil
	default:
		return nil, fmt.Errorf("Unsupported postgresql version: %s", version)
	}
}

var mdbPGSettingsFieldsInfo18 = newObjectFieldsInfo().
	addType(config.PostgresqlConfig18{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig18_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig18_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig18_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig18_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig18_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig18_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig18_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig18_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig18_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig18_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig18_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig18_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig18_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig18_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig18_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig18_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig18_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig18_PasswordEncryption_name,
		int(config.PostgresqlConfig18_PASSWORD_ENCRYPTION_SCRAM_SHA_256.Number()),
	).
	addEnumGeneratedNamesWithCompareAndValidFuncs("debug_parallel_query", config.PostgresqlConfig18_DebugParallelQuery_name).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig18_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo18_1C = newObjectFieldsInfo().
	addType(config.PostgresqlConfig18_1C{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig18_1C_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig18_1C_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig18_1C_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig18_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig18_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig18_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig18_1C_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig18_1C_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig18_1C_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig18_1C_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig18_1C_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig18_1C_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig18_1C_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig18_1C_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig18_1C_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig18_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig18_1C_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig18_1C_PasswordEncryption_name,
		int(config.PostgresqlConfig18_1C_PASSWORD_ENCRYPTION_SCRAM_SHA_256.Number()),
	).
	addEnumGeneratedNamesWithCompareAndValidFuncs("debug_parallel_query", config.PostgresqlConfig18_1C_DebugParallelQuery_name).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig18_1C_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo17 = newObjectFieldsInfo().
	addType(config.PostgresqlConfig17{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig17_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig17_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig17_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig17_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig17_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig17_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig17_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig17_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig17_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig17_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig17_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig17_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig17_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig17_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig17_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig17_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig17_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig17_PasswordEncryption_name,
		int(config.PostgresqlConfig17_PASSWORD_ENCRYPTION_SCRAM_SHA_256.Number()),
	).
	addEnumGeneratedNamesWithCompareAndValidFuncs("debug_parallel_query", config.PostgresqlConfig17_DebugParallelQuery_name).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig17_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo17_1C = newObjectFieldsInfo().
	addType(config.PostgresqlConfig17_1C{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig17_1C_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig17_1C_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig17_1C_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig17_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig17_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig17_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig17_1C_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig17_1C_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig17_1C_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig17_1C_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig17_1C_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig17_1C_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig17_1C_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig17_1C_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig17_1C_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig17_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig17_1C_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig17_1C_PasswordEncryption_name,
		int(config.PostgresqlConfig17_1C_PASSWORD_ENCRYPTION_SCRAM_SHA_256.Number()),
	).
	addEnumGeneratedNamesWithCompareAndValidFuncs("debug_parallel_query", config.PostgresqlConfig17_1C_DebugParallelQuery_name).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig17_1C_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo16 = newObjectFieldsInfo().
	addType(config.PostgresqlConfig16{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig16_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig16_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig16_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig16_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig16_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig16_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig16_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig16_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig16_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig16_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig16_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig16_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig16_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig16_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig16_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig16_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig16_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig16_PasswordEncryption_name,
		int(config.PostgresqlConfig16_PASSWORD_ENCRYPTION_MD5.Number()),
	).
	addEnumGeneratedNamesWithCompareAndValidFuncs("debug_parallel_query", config.PostgresqlConfig16_DebugParallelQuery_name).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig16_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo16_1C = newObjectFieldsInfo().
	addType(config.PostgresqlConfig16_1C{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig16_1C_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig16_1C_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig16_1C_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig16_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig16_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig16_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig16_1C_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig16_1C_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig16_1C_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig16_1C_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig16_1C_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig16_1C_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig16_1C_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig16_1C_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig16_1C_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig16_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig16_1C_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig16_1C_PasswordEncryption_name,
		int(config.PostgresqlConfig16_1C_PASSWORD_ENCRYPTION_MD5.Number()),
	).
	addEnumGeneratedNamesWithCompareAndValidFuncs("debug_parallel_query", config.PostgresqlConfig16_1C_DebugParallelQuery_name).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig16_1C_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo15 = newObjectFieldsInfo().
	addType(config.PostgresqlConfig15{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig15_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig15_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig15_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("force_parallel_mode", config.PostgresqlConfig15_ForceParallelMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig15_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig15_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig15_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig15_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig15_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig15_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig15_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig15_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig15_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig15_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig15_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig15_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig15_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig15_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig15_PasswordEncryption_name,
		int(config.PostgresqlConfig15_PASSWORD_ENCRYPTION_MD5.Number()),
	).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig15_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo15_1C = newObjectFieldsInfo().
	addType(config.PostgresqlConfig15_1C{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig15_1C_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig15_1C_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig15_1C_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("force_parallel_mode", config.PostgresqlConfig15_1C_ForceParallelMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig15_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig15_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig15_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig15_1C_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig15_1C_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig15_1C_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig15_1C_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig15_1C_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig15_1C_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig15_1C_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig15_1C_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig15_1C_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig15_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig15_1C_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig15_1C_PasswordEncryption_name,
		int(config.PostgresqlConfig15_1C_PASSWORD_ENCRYPTION_MD5.Number()),
	).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig15_1C_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo14 = newObjectFieldsInfo().
	addType(config.PostgresqlConfig14{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig14_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig14_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig14_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("force_parallel_mode", config.PostgresqlConfig14_ForceParallelMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig14_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig14_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig14_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig14_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig14_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig14_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig14_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig14_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig14_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig14_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig14_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig14_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig14_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig14_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig14_PasswordEncryption_name,
		int(config.PostgresqlConfig14_PASSWORD_ENCRYPTION_MD5.Number()),
	).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig14_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo14_1C = newObjectFieldsInfo().
	addType(config.PostgresqlConfig14_1C{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig14_1C_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig14_1C_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig14_1C_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("force_parallel_mode", config.PostgresqlConfig14_1C_ForceParallelMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig14_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig14_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig14_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig14_1C_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig14_1C_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig14_1C_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig14_1C_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig14_1C_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig14_1C_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig14_1C_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig14_1C_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig14_1C_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig14_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("auto_explain_log_format", config.PostgresqlConfig14_1C_AutoExplainLogFormat_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig14_1C_PasswordEncryption_name,
		int(config.PostgresqlConfig14_1C_PASSWORD_ENCRYPTION_MD5.Number()),
	).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig14_1C_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo13 = newObjectFieldsInfo().
	addType(config.PostgresqlConfig13{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig13_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig13_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig13_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("force_parallel_mode", config.PostgresqlConfig13_ForceParallelMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig13_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig13_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig13_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig13_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig13_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig13_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig13_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig13_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig13_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig13_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig13_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig13_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig13_LogLevel_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig13_PasswordEncryption_name,
		int(config.PostgresqlConfig13_PASSWORD_ENCRYPTION_MD5.Number()),
	).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig13_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)

var mdbPGSettingsFieldsInfo13_1C = newObjectFieldsInfo().
	addType(config.PostgresqlConfig13_1C{}, []reflect.Type{}).
	addEnumGeneratedNamesWithCompareAndValidFuncs("wal_level", config.PostgresqlConfig13_1C_WalLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("synchronous_commit", config.PostgresqlConfig13_1C_SynchronousCommit_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("constraint_exclusion", config.PostgresqlConfig13_1C_ConstraintExclusion_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("force_parallel_mode", config.PostgresqlConfig13_1C_ForceParallelMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("client_min_messages", config.PostgresqlConfig13_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_messages", config.PostgresqlConfig13_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_min_error_statement", config.PostgresqlConfig13_1C_LogLevel_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_error_verbosity", config.PostgresqlConfig13_1C_LogErrorVerbosity_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("log_statement", config.PostgresqlConfig13_1C_LogStatement_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("default_transaction_isolation", config.PostgresqlConfig13_1C_TransactionIsolation_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("bytea_output", config.PostgresqlConfig13_1C_ByteaOutput_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmlbinary", config.PostgresqlConfig13_1C_XmlBinary_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("xmloption", config.PostgresqlConfig13_1C_XmlOption_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("backslash_quote", config.PostgresqlConfig13_1C_BackslashQuote_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("plan_cache_mode", config.PostgresqlConfig13_1C_PlanCacheMode_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_debug_print", config.PostgresqlConfig13_1C_PgHintPlanDebugPrint_name).
	addEnumGeneratedNamesWithCompareAndValidFuncs("pg_hint_plan_message_level", config.PostgresqlConfig13_1C_LogLevel_name).
	addEnumGeneratedNamesWithDefaultValueCompareAndValidFuncs(
		"password_encryption",
		config.PostgresqlConfig13_1C_PasswordEncryption_name,
		int(config.PostgresqlConfig13_1C_PASSWORD_ENCRYPTION_MD5.Number()),
	).
	addSkipEnumGeneratedNames("shared_preload_libraries", config.PostgresqlConfig13_1C_SharedPreloadLibraries_name, defaultStringOfEnumsCheck("shared_preload_libraries"), defaultStringCompare)
