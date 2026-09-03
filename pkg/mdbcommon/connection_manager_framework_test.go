package mdbcommon

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// The attribute sits inside a config block, like it does in the cluster resources, so the cases
// cover a missing ancestor as well as a missing attribute: validation runs on every plan, including
// the ones that destroy the cluster or leave the block out entirely.
func TestValidateClusterConnectionManagerFromConfig(t *testing.T) {
	ctx := context.Background()

	cmType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"enabled":               tftypes.Bool,
			"connections_folder_id": tftypes.String,
			"secrets_folder_id":     tftypes.String,
		},
	}
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{"connection_manager": cmType},
	}
	clusterType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{"config": configType},
	}

	withConnectionManager := func(cm tftypes.Value) tftypes.Value {
		return tftypes.NewValue(configType, map[string]tftypes.Value{"connection_manager": cm})
	}
	withEnabled := func(enabled tftypes.Value) tftypes.Value {
		return withConnectionManager(tftypes.NewValue(cmType, map[string]tftypes.Value{
			"enabled":               enabled,
			"connections_folder_id": tftypes.NewValue(tftypes.String, nil),
			"secrets_folder_id":     tftypes.NewValue(tftypes.String, nil),
		}))
	}

	cases := []struct {
		name        string
		config      tftypes.Value
		expectError bool
	}{
		{"config block omitted", tftypes.NewValue(configType, nil), false},
		{"block omitted", withConnectionManager(tftypes.NewValue(cmType, nil)), false},
		{"block unknown", withConnectionManager(tftypes.NewValue(cmType, tftypes.UnknownValue)), false},
		{"enabled omitted", withEnabled(tftypes.NewValue(tftypes.Bool, nil)), false},
		{"enabled unknown", withEnabled(tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)), false},
		{"enabled true", withEnabled(tftypes.NewValue(tftypes.Bool, true)), false},
		{"enabled false", withEnabled(tftypes.NewValue(tftypes.Bool, false)), true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := tfsdk.Config{
				Raw: tftypes.NewValue(clusterType, map[string]tftypes.Value{"config": c.config}),
				Schema: schema.Schema{
					Blocks: map[string]schema.Block{
						"config": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"connection_manager": ClusterConnectionManagerFrameworkSchema(),
							},
						},
					},
				},
			}

			var diags diag.Diagnostics
			ValidateClusterConnectionManagerFromConfig(ctx, cfg, path.Root("config").AtName("connection_manager"), &diags)

			if diags.HasError() != c.expectError {
				t.Errorf("expected error: %t, got diagnostics: %v", c.expectError, diags)
			}
		})
	}
}
