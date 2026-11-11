package trino_cluster

import (
	"fmt"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
)

type ResourceGroups struct {
	RootGroups     []ResourceGroup `json:"rootGroups"`
	Selectors      []SelectorRule  `json:"selectors"`
	CpuQuotaPeriod string          `json:"cpuQuotaPeriod"`
}

func (r *ResourceGroups) Equal(other *ResourceGroups) bool {
	if r == nil && other == nil {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	if r.CpuQuotaPeriod != other.CpuQuotaPeriod {
		return false
	}
	if len(r.RootGroups) != len(other.RootGroups) {
		return false
	}
	for i := range r.RootGroups {
		if !r.RootGroups[i].Equal(&other.RootGroups[i]) {
			return false
		}
	}
	if len(r.Selectors) != len(other.Selectors) {
		return false
	}
	for i := range r.Selectors {
		if !r.Selectors[i].Equal(&other.Selectors[i]) {
			return false
		}
	}
	return true
}

func (r *ResourceGroups) Validate() error {
	for i := range r.RootGroups {
		if err := r.RootGroups[i].Validate(); err != nil {
			return fmt.Errorf("rootGroups[%d]: %w", i, err)
		}
	}
	for i := range r.Selectors {
		if err := r.Selectors[i].Validate(); err != nil {
			return fmt.Errorf("selectors[%d]: %w", i, err)
		}
	}
	return nil
}

func (r *ResourceGroups) ToAPI() *trino.ResourceGroupsConfig {
	rootGroups := make([]*trino.ResourceGroupConfig, 0, len(r.RootGroups))
	for _, rootGroup := range r.RootGroups {
		rootGroups = append(rootGroups, rootGroup.ToAPI())
	}

	selectors := make([]*trino.SelectorRuleConfig, 0, len(r.Selectors))
	for _, selector := range r.Selectors {
		selectors = append(selectors, selector.ToAPI())
	}

	return &trino.ResourceGroupsConfig{
		RootGroups:     rootGroups,
		Selectors:      selectors,
		CpuQuotaPeriod: r.CpuQuotaPeriod,
	}
}

func ResourceGroupsFromAPI(api *trino.ResourceGroupsConfig) *ResourceGroups {
	if api == nil {
		return nil
	}

	rootGroups := make([]ResourceGroup, 0, len(api.RootGroups))
	for _, rootGroup := range api.RootGroups {
		rootGroups = append(rootGroups, resourceGroupFromAPI(rootGroup))
	}

	selectors := make([]SelectorRule, 0, len(api.Selectors))
	for _, selector := range api.Selectors {
		selectors = append(selectors, selectorRuleFromAPI(selector))
	}

	return &ResourceGroups{
		RootGroups:     rootGroups,
		Selectors:      selectors,
		CpuQuotaPeriod: api.CpuQuotaPeriod,
	}
}

type ResourceGroup struct {
	Name                 string          `json:"name"`
	MaxQueued            int64           `json:"maxQueued"`
	SoftConcurrencyLimit int64           `json:"softConcurrencyLimit"`
	HardConcurrencyLimit int64           `json:"hardConcurrencyLimit"`
	SoftMemoryLimit      string          `json:"softMemoryLimit"`
	SoftCpuLimit         string          `json:"softCpuLimit"`
	HardCpuLimit         string          `json:"hardCpuLimit"`
	SchedulingPolicy     string          `json:"schedulingPolicy"`
	SchedulingWeight     int64           `json:"schedulingWeight"`
	SubGroups            []ResourceGroup `json:"subGroups"`
}

func (r *ResourceGroup) Equal(other *ResourceGroup) bool {
	if r == nil && other == nil {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	if r.Name != other.Name ||
		r.MaxQueued != other.MaxQueued ||
		r.SoftConcurrencyLimit != other.SoftConcurrencyLimit ||
		r.HardConcurrencyLimit != other.HardConcurrencyLimit ||
		r.SoftMemoryLimit != other.SoftMemoryLimit ||
		r.SoftCpuLimit != other.SoftCpuLimit ||
		r.HardCpuLimit != other.HardCpuLimit ||
		schedulingPolicyToAPI(r.SchedulingPolicy) != schedulingPolicyToAPI(other.SchedulingPolicy) ||
		r.SchedulingWeight != other.SchedulingWeight {
		return false
	}
	if len(r.SubGroups) != len(other.SubGroups) {
		return false
	}
	for i := range r.SubGroups {
		if !r.SubGroups[i].Equal(&other.SubGroups[i]) {
			return false
		}
	}
	return true
}

func (r *ResourceGroup) Validate() error {
	if r.SchedulingPolicy != "" && schedulingPolicyToAPI(r.SchedulingPolicy) == trino.ResourceGroupConfig_SCHEDULING_POLICY_UNSPECIFIED {
		return fmt.Errorf("invalid schedulingPolicy: %q", r.SchedulingPolicy)
	}
	for i := range r.SubGroups {
		if err := r.SubGroups[i].Validate(); err != nil {
			return fmt.Errorf("subGroups[%d]: %w", i, err)
		}
	}
	return nil
}

func (r *ResourceGroup) ToAPI() *trino.ResourceGroupConfig {
	subGroups := make([]*trino.ResourceGroupConfig, 0, len(r.SubGroups))
	for _, subGroup := range r.SubGroups {
		subGroups = append(subGroups, subGroup.ToAPI())
	}

	return &trino.ResourceGroupConfig{
		Name:                 r.Name,
		MaxQueued:            r.MaxQueued,
		SoftConcurrencyLimit: r.SoftConcurrencyLimit,
		HardConcurrencyLimit: r.HardConcurrencyLimit,
		SoftMemoryLimit:      r.SoftMemoryLimit,
		SoftCpuLimit:         r.SoftCpuLimit,
		HardCpuLimit:         r.HardCpuLimit,
		SchedulingPolicy:     schedulingPolicyToAPI(r.SchedulingPolicy),
		SchedulingWeight:     r.SchedulingWeight,
		SubGroups:            subGroups,
	}
}

func resourceGroupFromAPI(api *trino.ResourceGroupConfig) ResourceGroup {
	subGroups := make([]ResourceGroup, 0, len(api.SubGroups))
	for _, subGroup := range api.SubGroups {
		subGroups = append(subGroups, resourceGroupFromAPI(subGroup))
	}

	return ResourceGroup{
		Name:                 api.Name,
		MaxQueued:            api.MaxQueued,
		SoftConcurrencyLimit: api.SoftConcurrencyLimit,
		HardConcurrencyLimit: api.HardConcurrencyLimit,
		SoftMemoryLimit:      api.SoftMemoryLimit,
		SoftCpuLimit:         api.SoftCpuLimit,
		HardCpuLimit:         api.HardCpuLimit,
		SchedulingPolicy:     schedulingPolicyFromAPI(api.SchedulingPolicy),
		SchedulingWeight:     api.SchedulingWeight,
		SubGroups:            subGroups,
	}
}

func schedulingPolicyToAPI(schedulingPolicy string) trino.ResourceGroupConfig_SchedulingPolicy {
	switch strings.ToLower(schedulingPolicy) {
	case "fair":
		return trino.ResourceGroupConfig_FAIR
	case "weighted":
		return trino.ResourceGroupConfig_WEIGHTED
	case "weighted_fair":
		return trino.ResourceGroupConfig_WEIGHTED_FAIR
	case "query_priority":
		return trino.ResourceGroupConfig_QUERY_PRIORITY
	default:
		return trino.ResourceGroupConfig_SCHEDULING_POLICY_UNSPECIFIED
	}
}

func schedulingPolicyFromAPI(schedulingPolicy trino.ResourceGroupConfig_SchedulingPolicy) string {
	if schedulingPolicy == trino.ResourceGroupConfig_SCHEDULING_POLICY_UNSPECIFIED {
		return ""
	}
	return strings.ToLower(trino.ResourceGroupConfig_SchedulingPolicy_name[int32(schedulingPolicy)])
}

type SelectorRule struct {
	User       string   `json:"user"`
	UserGroup  string   `json:"userGroup"`
	Source     string   `json:"source"`
	QueryType  string   `json:"queryType"`
	ClientTags []string `json:"clientTags"`
	Group      string   `json:"group"`
}

func (r *SelectorRule) Equal(other *SelectorRule) bool {
	if r == nil && other == nil {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	if r.User != other.User ||
		r.UserGroup != other.UserGroup ||
		r.Source != other.Source ||
		r.QueryType != other.QueryType ||
		queryTypeToAPI(r.QueryType) != queryTypeToAPI(other.QueryType) ||
		r.Group != other.Group {
		return false
	}
	if len(r.ClientTags) != len(other.ClientTags) {
		return false
	}
	for i := range r.ClientTags {
		if r.ClientTags[i] != other.ClientTags[i] {
			return false
		}
	}
	return true
}

func (r *SelectorRule) Validate() error {
	if r.QueryType != "" && queryTypeToAPI(r.QueryType) == trino.SelectorRuleConfig_QUERY_TYPE_UNSPECIFIED {
		return fmt.Errorf("invalid queryType: %q", r.QueryType)
	}
	return nil
}

func (r *SelectorRule) ToAPI() *trino.SelectorRuleConfig {
	return &trino.SelectorRuleConfig{
		User:       r.User,
		UserGroup:  r.UserGroup,
		Source:     r.Source,
		QueryType:  queryTypeToAPI(r.QueryType),
		ClientTags: r.ClientTags,
		Group:      r.Group,
	}
}

func selectorRuleFromAPI(api *trino.SelectorRuleConfig) SelectorRule {
	return SelectorRule{
		User:       api.User,
		UserGroup:  api.UserGroup,
		Source:     api.Source,
		QueryType:  queryTypeFromAPI(api.QueryType),
		ClientTags: api.ClientTags,
		Group:      api.Group,
	}
}

func queryTypeToAPI(queryType string) trino.SelectorRuleConfig_QueryType {
	switch strings.ToUpper(queryType) {
	case "SELECT":
		return trino.SelectorRuleConfig_SELECT
	case "EXPLAIN":
		return trino.SelectorRuleConfig_EXPLAIN
	case "DESCRIBE":
		return trino.SelectorRuleConfig_DESCRIBE
	case "INSERT":
		return trino.SelectorRuleConfig_INSERT
	case "UPDATE":
		return trino.SelectorRuleConfig_UPDATE
	case "DELETE":
		return trino.SelectorRuleConfig_DELETE
	case "MERGE":
		return trino.SelectorRuleConfig_MERGE
	case "ANALYZE":
		return trino.SelectorRuleConfig_ANALYZE
	case "DATA_DEFINITION":
		return trino.SelectorRuleConfig_DATA_DEFINITION
	case "ALTER_TABLE_EXECUTE":
		return trino.SelectorRuleConfig_ALTER_TABLE_EXECUTE
	default:
		return trino.SelectorRuleConfig_QUERY_TYPE_UNSPECIFIED
	}
}

func queryTypeFromAPI(queryType trino.SelectorRuleConfig_QueryType) string {
	if queryType == trino.SelectorRuleConfig_QUERY_TYPE_UNSPECIFIED {
		return ""
	}
	return strings.ToUpper(trino.SelectorRuleConfig_QueryType_name[int32(queryType)])
}
