package statecheck

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraform2 "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func CreateResourceData(t *testing.T, schemaObject map[string]*schema.Schema, rawInitialState map[string]interface{},
	diffAttributes map[string]*terraform2.ResourceAttrDiff,
) *schema.ResourceData {
	t.Helper()
	ctx := context.Background()
	internalMap := schema.InternalMap(schemaObject)

	emptyState := terraform2.NewInstanceStateShimmedFromValue(cty.ObjectVal(map[string]cty.Value{}), 1)
	initialDiff, err := internalMap.Diff(ctx, emptyState.DeepCopy(), terraform2.NewResourceConfigRaw(rawInitialState), nil, nil, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	targetDiff := &terraform2.InstanceDiff{Attributes: diffAttributes}

	resourceData, err := internalMap.Data(emptyState.MergeDiff(initialDiff), targetDiff)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return resourceData
}
