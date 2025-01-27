package mdb_opensearch_cluster_test

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	pc "github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var _ pc.PlanCheck = expectNoChangesAt{}

type expectNoChangesAt struct {
	resourceAddress string
	attributePath   tfjsonpath.Path
}

// CheckPlan implements the plan check logic.
func (e expectNoChangesAt) CheckPlan(ctx context.Context, req pc.CheckPlanRequest, resp *pc.CheckPlanResponse) {
	var change *tfjson.ResourceChange

	for _, rc := range req.Plan.ResourceChanges {
		if e.resourceAddress == rc.Address {
			change = rc
			break
		}
	}

	if change == nil {
		//No changes in plan for selected output, so return
		return
	}

	before, err := tfjsonpath.Traverse(change.Change.Before, e.attributePath)
	if err != nil {
		resp.Error = fmt.Errorf("Failed to get attr value before changes: %s", err)
		return
	}

	after, err := tfjsonpath.Traverse(change.Change.After, e.attributePath)
	if err != nil {
		resp.Error = fmt.Errorf("Failed to get attr value after changes: %s", err)
		return
	}

	err = compare.ValuesSame().CompareValues(before, after)
	if err == nil {
		return
	}

	resp.Error = fmt.Errorf("%s - Resource has attribute in plan: %s", e.resourceAddress, err)
}

// ExpectNoChangesAt returns a plan check that asserts that there are no changes
// at the given resource.
func ExpectNoChangesAt(resourceAddress string, attributePath tfjsonpath.Path) pc.PlanCheck {
	return expectNoChangesAt{
		resourceAddress: resourceAddress,
		attributePath:   attributePath,
	}
}
