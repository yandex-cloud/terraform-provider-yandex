package helm_release

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	marketplace "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/marketplace/v1"
)

func TestUserValuesFromPlan(t *testing.T) {
	tests := map[string]struct {
		plan   helmReleaseResourceModel
		values []*marketplace.ValueWithKey
	}{
		"as keys": {
			plan: helmReleaseResourceModel{
				Name:      types.StringValue("app-a"),
				Namespace: types.StringValue("ns-a"),
			},
			values: []*marketplace.ValueWithKey{
				{
					Key: nameValue,
					Value: &marketplace.Value{
						Value: &marketplace.Value_TypedValue{
							TypedValue: "app-a",
						},
					},
				},
				{
					Key: namespaceValue,
					Value: &marketplace.Value{
						Value: &marketplace.Value_TypedValue{
							TypedValue: "ns-a",
						},
					},
				},
			},
		},
		"as values": {
			plan: helmReleaseResourceModel{
				UserValues: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"applicationName": types.StringValue("app-b"),
						"namespace":       types.StringValue("ns-b"),
					},
				),
			},
			values: []*marketplace.ValueWithKey{
				{
					Key: nameValue,
					Value: &marketplace.Value{
						Value: &marketplace.Value_TypedValue{
							TypedValue: "app-b",
						},
					},
				},
				{
					Key: namespaceValue,
					Value: &marketplace.Value{
						Value: &marketplace.Value_TypedValue{
							TypedValue: "ns-b",
						},
					},
				},
			},
		},
		"override": {
			plan: helmReleaseResourceModel{
				Name:      types.StringValue("app-a"),
				Namespace: types.StringValue("ns-a"),
				UserValues: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"applicationName": types.StringValue("app-b"),
					},
				),
			},
			values: []*marketplace.ValueWithKey{
				{
					Key: nameValue,
					Value: &marketplace.Value{
						Value: &marketplace.Value_TypedValue{
							TypedValue: "app-b",
						},
					},
				},
				{
					Key: namespaceValue,
					Value: &marketplace.Value{
						Value: &marketplace.Value_TypedValue{
							TypedValue: "ns-a",
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			values, _ := userValuesFromPlan(ctx, tc.plan)
			if !valuesEqual(values, tc.values) {
				t.Errorf("expected %v, got %v", tc.values, values)
			}
		})
	}
}

func valuesEqual(values []*marketplace.ValueWithKey, expected []*marketplace.ValueWithKey) bool {
	if len(values) != len(expected) {
		return false
	}

loop:
	for _, v := range values {
		for _, ex := range expected {
			if v.Key == ex.Key {
				if v.Value.GetTypedValue() != ex.Value.GetTypedValue() {
					return false
				}
				continue loop
			}
		}
		return false
	}
	return true
}
