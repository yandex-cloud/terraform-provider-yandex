package yandex_cloudregistry_lifecycle_policy_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	yandexframework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func TestResourceIDValidation(t *testing.T) {
	ctx := context.Background()
	server := providerserver.NewProtocol6(yandexframework.NewFrameworkProvider())()
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schemas.Diagnostics)
	resourceSchema := schemas.ResourceSchemas["yandex_cloudregistry_lifecycle_policy"]
	require.NotNil(t, resourceSchema)
	objectType := resourceSchema.ValueType().(tftypes.Object)

	for _, tc := range []struct {
		name      string
		id        interface{}
		wantError bool
	}{
		{name: "without_id"},
		{name: "with_id", id: "manually-assigned-id", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := make(map[string]tftypes.Value)
			for name, typ := range objectType.AttributeTypes {
				values[name] = tftypes.NewValue(typ, nil)
			}
			values["name"] = tftypes.NewValue(tftypes.String, "id-validation-test")
			values["registry_id"] = tftypes.NewValue(tftypes.String, "parent-registry-id")
			values["policy_id"] = tftypes.NewValue(tftypes.String, tc.id)
			config, err := tfprotov6.NewDynamicValue(objectType, tftypes.NewValue(objectType, values))
			require.NoError(t, err)
			response, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{
				TypeName: "yandex_cloudregistry_lifecycle_policy",
				Config:   &config,
			})
			require.NoError(t, err)
			if !tc.wantError {
				require.Empty(t, response.Diagnostics)
				return
			}
			require.Len(t, response.Diagnostics, 1)
			require.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)
			require.Equal(t, "Invalid Configuration for Read-Only Attribute", response.Diagnostics[0].Summary)
			require.Equal(t, tftypes.NewAttributePath().WithAttributeName("policy_id"), response.Diagnostics[0].Attribute)
		})
	}
}
