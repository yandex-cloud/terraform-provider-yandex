package mdb_redis_cluster_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func MapWarningHostsChangedAfterImport() planmodifier.Map {
	return warningOnChangeHosts{}
}

type warningOnChangeHosts struct{}

func (m warningOnChangeHosts) Description(_ context.Context) string {
	return "Add warnings if change plan wrong with added and deleted hosts."
}

func (m warningOnChangeHosts) MarkdownDescription(_ context.Context) string {
	return "Add warnings if change plan wrong with added and deleted hosts."
}

func (m warningOnChangeHosts) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	stateHostsMap := make(map[string]Host)
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &stateHostsMap, false)...)
	planHostsMap := make(map[string]Host)
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &planHostsMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapping := make(map[string]string)
	fixedPlan := make(map[string]Host)
	usedState := make(map[string]struct{})
	for label := range stateHostsMap {
		if _, ok := planHostsMap[label]; ok {
			fixedPlan[label] = planHostsMap[label]
			usedState[label] = struct{}{}
		}
	}

	//fully match
	for label, stateHost := range stateHostsMap {
		for planLabel, planHost := range planHostsMap {
			_, okState := fixedPlan[planLabel]
			_, okPlan := usedState[label]
			if okState || okPlan {
				continue
			}
			if redisHostService.FullyMatch(planHost, stateHost) {
				fixedPlan[planLabel] = stateHost
				usedState[label] = struct{}{}
				mapping[label] = planLabel
			}
		}
	}

	//partitial match
	for label, stateHost := range stateHostsMap {
		for planLabel, planHost := range planHostsMap {
			_, okState := fixedPlan[planLabel]
			_, okPlan := usedState[label]
			if okState || okPlan {
				continue
			}
			if redisHostService.PartialMatch(planHost, stateHost) {
				fixedPlan[planLabel] = stateHost
				usedState[label] = struct{}{}
				mapping[label] = planLabel
			}
		}
	}
	if len(mapping) > 0 {
		warn := ""
		for stateLabel, planLabel := range mapping {
			warn += fmt.Sprintf("Host with the label %q will change the label to %q, without any opertations\n", stateLabel, planLabel)
		}
		resp.Diagnostics.AddWarning(
			"Wrong plan",
			warn,
		)
	}
}
