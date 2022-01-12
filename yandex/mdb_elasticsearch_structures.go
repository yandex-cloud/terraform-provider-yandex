package yandex

import (
	"context"
	"errors"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/elasticsearch/v1"
)

func parseElasticsearchEnv(e string) (elasticsearch.Cluster_Environment, error) {
	v, ok := elasticsearch.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(elasticsearch.Cluster_Environment_value)), e)
	}
	return elasticsearch.Cluster_Environment(v), nil
}

func parseElasticsearchHostType(t string) (elasticsearch.Host_Type, error) {
	v, ok := elasticsearch.Host_Type_value[t]
	if !ok {
		return 0, fmt.Errorf("value for 'host.type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(elasticsearch.Host_Type_value)), t)
	}
	return elasticsearch.Host_Type(v), nil
}

func expandElasticsearchConfigSpec(d *schema.ResourceData) *elasticsearch.ConfigSpec {
	config := &elasticsearch.ConfigSpec{

		Version: d.Get("config.0.version").(string),

		Edition: d.Get("config.0.edition").(string),

		AdminPassword: d.Get("config.0.admin_password").(string),

		ElasticsearchSpec: &elasticsearch.ElasticsearchSpec{
			DataNode: &elasticsearch.ElasticsearchSpec_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: d.Get("config.0.data_node.0.resources.0.resource_preset_id").(string),
					DiskTypeId:       d.Get("config.0.data_node.0.resources.0.disk_type_id").(string),
					DiskSize:         toBytes(d.Get("config.0.data_node.0.resources.0.disk_size").(int)),
				},
			},
			Plugins: convertStringSet(d.Get("config.0.plugins").(*schema.Set)),
		},
	}

	if _, exist := d.GetOk("config.0.master_node"); exist {
		config.ElasticsearchSpec.MasterNode = &elasticsearch.ElasticsearchSpec_MasterNode{
			Resources: &elasticsearch.Resources{
				ResourcePresetId: d.Get("config.0.master_node.0.resources.0.resource_preset_id").(string),
				DiskTypeId:       d.Get("config.0.master_node.0.resources.0.disk_type_id").(string),
				DiskSize:         toBytes(d.Get("config.0.master_node.0.resources.0.disk_size").(int)),
			},
		}
	}

	return config
}

func expandElasticsearchConfigSpecUpdate(d *schema.ResourceData) *elasticsearch.ConfigSpecUpdate {
	config := &elasticsearch.ConfigSpecUpdate{

		Version: d.Get("config.0.version").(string),

		Edition: d.Get("config.0.edition").(string),

		AdminPassword: d.Get("config.0.admin_password").(string),

		ElasticsearchSpec: &elasticsearch.ElasticsearchSpec{
			DataNode: &elasticsearch.ElasticsearchSpec_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: d.Get("config.0.data_node.0.resources.0.resource_preset_id").(string),
					DiskTypeId:       d.Get("config.0.data_node.0.resources.0.disk_type_id").(string),
					DiskSize:         toBytes(d.Get("config.0.data_node.0.resources.0.disk_size").(int)),
				},
			},
			Plugins: convertStringSet(d.Get("config.0.plugins").(*schema.Set)),
		},
	}

	if _, exist := d.GetOk("config.0.master_node"); exist {
		config.ElasticsearchSpec.MasterNode = &elasticsearch.ElasticsearchSpec_MasterNode{
			Resources: &elasticsearch.Resources{
				ResourcePresetId: d.Get("config.0.master_node.0.resources.0.resource_preset_id").(string),
				DiskTypeId:       d.Get("config.0.master_node.0.resources.0.disk_type_id").(string),
				DiskSize:         toBytes(d.Get("config.0.master_node.0.resources.0.disk_size").(int)),
			},
		}
	}

	return config
}

func flattenElasticsearchClusterConfig(config *elasticsearch.ClusterConfig, password string) []interface{} {
	res := map[string]interface{}{
		"version":        config.Version,
		"edition":        config.Edition,
		"admin_password": password,
		"plugins":        config.Elasticsearch.Plugins,
		"data_node": []interface{}{map[string]interface{}{
			"resources": []interface{}{map[string]interface{}{
				"resource_preset_id": config.Elasticsearch.DataNode.Resources.ResourcePresetId,
				"disk_type_id":       config.Elasticsearch.DataNode.Resources.DiskTypeId,
				"disk_size":          toGigabytes(config.Elasticsearch.DataNode.Resources.DiskSize),
			}},
		}},
	}

	if config.Elasticsearch.MasterNode != nil && config.Elasticsearch.MasterNode.Resources != nil {
		res["master_node"] = []interface{}{map[string]interface{}{
			"resources": []interface{}{map[string]interface{}{
				"resource_preset_id": config.Elasticsearch.MasterNode.Resources.ResourcePresetId,
				"disk_type_id":       config.Elasticsearch.MasterNode.Resources.DiskTypeId,
				"disk_size":          toGigabytes(config.Elasticsearch.MasterNode.Resources.DiskSize),
			}},
		}}
	}

	return []interface{}{res}
}

type ElasticsearchHost struct {
	Name     string
	Type     elasticsearch.Host_Type
	Fqdn     string
	Zone     string
	Subnet   string
	PublicIp bool
}

func expandElasticsearchHosts(data interface{}) (ElasticsearchHostList, error) {
	if data == nil {
		return nil, nil
	}

	var result ElasticsearchHostList
	hosts, ok := data.(*schema.Set)
	if !ok {
		return result, nil
	}

	for _, v := range hosts.List() {
		config := v.(map[string]interface{})
		host, err := expandElasticsearchHost(config)
		if err != nil {
			return nil, err
		}
		result = append(result, host)
	}

	return result, nil
}

func expandElasticsearchHost(config map[string]interface{}) (*ElasticsearchHost, error) {
	host := &ElasticsearchHost{}

	if v, ok := config["name"]; ok {
		host.Name = v.(string)
	}

	if v, ok := config["type"]; ok {
		t, err := parseElasticsearchHostType(v.(string))
		if err != nil {
			return nil, err
		}
		host.Type = t
	}

	if v, ok := config["fqdn"]; ok {
		host.Fqdn = v.(string)
	}

	if v, ok := config["zone"]; ok {
		host.Zone = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		host.Subnet = v.(string)
	}

	if v, ok := config["assign_public_ip"]; ok {
		host.PublicIp = v.(bool)
	}

	return host, nil
}

func flattenElasticsearchHosts(hosts []*ElasticsearchHost) ([]interface{}, error) {
	res := []interface{}{}

	for _, h := range hosts {
		res = append(res, map[string]interface{}{
			"type":             h.Type.String(),
			"zone":             h.Zone,
			"subnet_id":        h.Subnet,
			"assign_public_ip": h.PublicIp,
			"fqdn":             h.Fqdn,
			"name":             h.Name,
		})
	}

	return res, nil
}

func convertElasticsearchHostsToSpecs(hosts []*ElasticsearchHost) []*elasticsearch.HostSpec {
	result := make([]*elasticsearch.HostSpec, len(hosts))
	for i, h := range hosts {
		result[i] = &elasticsearch.HostSpec{
			Type:           h.Type,
			ZoneId:         h.Zone,
			SubnetId:       h.Subnet,
			AssignPublicIp: h.PublicIp,
		}
	}
	return result
}

func convertElasticsearchActualHosts(hosts []*elasticsearch.Host) ElasticsearchHostList {
	result := make(ElasticsearchHostList, len(hosts))
	for i, host := range hosts {
		sn := host.SubnetId
		// special value, make difference between must be calculated and not defined (for porto)
		if sn == "" {
			sn = "none"
		}
		result[i] = &ElasticsearchHost{
			Name:     "", // map it later from state
			Type:     host.Type,
			Fqdn:     host.Name,
			Zone:     host.ZoneId,
			Subnet:   sn,
			PublicIp: host.AssignPublicIp,
		}
	}
	return result
}

// fill calculated fileds from other host
func (h *ElasticsearchHost) fill(o *ElasticsearchHost) {
	h.Fqdn = o.Fqdn
	if h.Subnet == "" {
		h.Subnet = o.Subnet
	}
}

// match checks that this host match config pattern
func (h *ElasticsearchHost) match(p *ElasticsearchHost) bool {
	return p.Zone == h.Zone && p.Type == h.Type && p.PublicIp == h.PublicIp && (p.Subnet == "" || p.Subnet == h.Subnet)
}

type ElasticsearchHostList []*ElasticsearchHost

func (l *ElasticsearchHostList) RemoveByFQDN(fqdn string) (*ElasticsearchHost, bool) {
	return l.RemoveBy(func(h *ElasticsearchHost) bool { return h.Fqdn == fqdn })
}

func (l *ElasticsearchHostList) RemoveByName(name string) (*ElasticsearchHost, bool) {
	return l.RemoveBy(func(h *ElasticsearchHost) bool { return h.Name == name })
}

func (l ElasticsearchHostList) HasMasters() bool {
	for i := range l {
		if l[i].Type == elasticsearch.Host_MASTER_NODE {
			return true
		}
	}
	return false
}

func (l ElasticsearchHostList) CountMasters() int {
	var c = 0
	for i := range l {
		if l[i].Type == elasticsearch.Host_MASTER_NODE {
			c++
		}
	}
	return c
}

type ElasticsearchHostPredicateFunc func(h *ElasticsearchHost) bool

func (l *ElasticsearchHostList) RemoveBy(p ElasticsearchHostPredicateFunc) (*ElasticsearchHost, bool) {
	if idx, ok := l.find(p); ok {
		return l.remove(idx), true
	}
	return nil, false
}

func (l ElasticsearchHostList) find(p ElasticsearchHostPredicateFunc) (int, bool) {
	for i := range l {
		if p(l[i]) {
			return i, true
		}
	}
	return 0, false
}

func (l *ElasticsearchHostList) remove(idx int) *ElasticsearchHost {
	s := *l
	last := len(s) - 1
	host := s[idx]

	s[idx] = s[last]
	*l = s[:last]

	return host
}

// Maps actually existing hosts with stored state. order doesn't matter.
func mapElasticsearchHostNames(actual []*elasticsearch.Host, state ElasticsearchHostList) []*ElasticsearchHost {
	var result = convertElasticsearchActualHosts(actual)

	var ready = make([]bool, len(result))

	// match host name by fqdn first
	for i := range result {
		if s, ok := state.RemoveByFQDN(result[i].Fqdn); ok {
			result[i].Name = s.Name
			ready[i] = true
		}
	}

	// match non named hosts from api with hosts in state, assign fqdn to names
	// matching in bipartie graph: we just uses greedy algo for our simple case

	// first hosts with a defined subnet, then all the others.
	for _, anySubnet := range []bool{false, true} {
		for i := range result {
			if !ready[i] {
				s, ok := state.RemoveBy(func(h *ElasticsearchHost) bool {
					return h.Fqdn == "" && (anySubnet || h.Subnet != "") && result[i].match(h)
				})
				if ok {
					result[i].Name = s.Name
					ready[i] = true
				}
			}
		}
	}

	return result
}

func elasticsearchHostDiffCustomize(ctx context.Context, rdiff *schema.ResourceDiff, _ interface{}) error {
	os, ns := rdiff.GetChange("host")

	oldHosts, _ := expandElasticsearchHosts(os)
	newHosts, _ := expandElasticsearchHosts(ns)

	if len(oldHosts) > 0 && oldHosts.CountMasters() != newHosts.CountMasters() {
		return errors.New("Adding/removing master nodes not supported")
	}

	var ready = make([]bool, len(newHosts))

	for i := range newHosts {
		if h, ok := oldHosts.RemoveByName(newHosts[i].Name); ok {
			// if not match, then recreate host with new calculated params (host changes not implemented)
			if h.match(newHosts[i]) {
				newHosts[i].fill(h)
			}
			ready[i] = true
		}
	}

	// drift detection: match non named hosts in state(old) with hosts from config(new)
	// matching in bipartie graph:  we just uses greedy algo for our simple case

	// first hosts with a defined subnet, then all the others.
	for _, anySubnet := range []bool{false, true} {
		for i := range newHosts {
			if !ready[i] && (anySubnet || newHosts[i].Subnet != "") {
				h, ok := oldHosts.RemoveBy(func(h *ElasticsearchHost) bool { return h.Name == "" && h.match(newHosts[i]) })
				if ok {
					newHosts[i].fill(h)
					ready[i] = true
				}
			}
		}
	}

	hs, err := flattenElasticsearchHosts(newHosts)
	if err != nil {
		return err
	}

	return rdiff.SetNew("host", hs)
}

func elasticsearchHostFQDNHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["fqdn"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func parseElasticsearchWeekDay(wd string) (elasticsearch.WeeklyMaintenanceWindow_WeekDay, error) {
	val, ok := elasticsearch.WeeklyMaintenanceWindow_WeekDay_value[wd]
	// do not allow WEEK_DAY_UNSPECIFIED
	if !ok || val == 0 {
		return elasticsearch.WeeklyMaintenanceWindow_WEEK_DAY_UNSPECIFIED,
			fmt.Errorf("value for 'day' should be one of %s, not `%s`",
				getJoinedKeys(getEnumValueMapKeysExt(elasticsearch.WeeklyMaintenanceWindow_WeekDay_value, true)), wd)
	}

	return elasticsearch.WeeklyMaintenanceWindow_WeekDay(val), nil
}

func expandElasticsearchMaintenanceWindow(d *schema.ResourceData) (*elasticsearch.MaintenanceWindow, error) {
	mwType, ok := d.GetOk("maintenance_window.0.type")
	if !ok {
		return nil, nil
	}

	result := &elasticsearch.MaintenanceWindow{}

	switch mwType {
	case "ANYTIME":
		timeSet := false
		if _, ok := d.GetOk("maintenance_window.0.day"); ok {
			timeSet = true
		}
		if _, ok := d.GetOk("maintenance_window.0.hour"); ok {
			timeSet = true
		}
		if timeSet {
			return nil, fmt.Errorf("with ANYTIME type of maintenance window both DAY and HOUR should be omitted")
		}
		result.SetAnytime(&elasticsearch.AnytimeMaintenanceWindow{})

	case "WEEKLY":
		weekly := &elasticsearch.WeeklyMaintenanceWindow{}
		if val, ok := d.GetOk("maintenance_window.0.day"); ok {
			var err error
			weekly.Day, err = parseElasticsearchWeekDay(val.(string))
			if err != nil {
				return nil, err
			}
		}
		if v, ok := d.GetOk("maintenance_window.0.hour"); ok {
			weekly.Hour = int64(v.(int))
		}

		result.SetWeeklyMaintenanceWindow(weekly)
	}

	return result, nil
}

func flattenElasticsearchMaintenanceWindow(mw *elasticsearch.MaintenanceWindow) []map[string]interface{} {
	result := map[string]interface{}{}

	if val := mw.GetAnytime(); val != nil {
		result["type"] = "ANYTIME"
	}

	if val := mw.GetWeeklyMaintenanceWindow(); val != nil {
		result["type"] = "WEEKLY"
		result["day"] = val.Day.String()
		result["hour"] = val.Hour
	}

	return []map[string]interface{}{result}
}
