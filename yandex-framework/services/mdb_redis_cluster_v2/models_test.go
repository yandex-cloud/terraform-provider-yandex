package mdb_redis_cluster_v2

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"
)

func TestClusterModelsMatchSchemas(t *testing.T) {
	ctx := context.Background()

	var resourceResponse resource.SchemaResponse
	NewResource().Schema(ctx, resource.SchemaRequest{}, &resourceResponse)
	if resourceResponse.Diagnostics.HasError() {
		t.Fatalf("resource schema diagnostics: %#v", resourceResponse.Diagnostics)
	}
	assertModelMatchesAttributes(t, Cluster{}, resourceResponse.Schema.Attributes)
	resourceConfigType, ok := resourceResponse.Schema.Attributes["config"].GetType().(types.ObjectType)
	if !ok {
		t.Fatalf("resource config type = %T, want types.ObjectType", resourceResponse.Schema.Attributes["config"].GetType())
	}
	assertModelMatchesAttributes(t, Config{}, resourceConfigType.AttrTypes)

	var dataSourceResponse datasource.SchemaResponse
	NewDataSource().Schema(ctx, datasource.SchemaRequest{}, &dataSourceResponse)
	if dataSourceResponse.Diagnostics.HasError() {
		t.Fatalf("data source schema diagnostics: %#v", dataSourceResponse.Diagnostics)
	}
	assertModelMatchesAttributes(t, dataSourceCluster{}, dataSourceResponse.Schema.Attributes)
	dataSourceConfigType, ok := dataSourceResponse.Schema.Attributes["config"].GetType().(types.ObjectType)
	if !ok {
		t.Fatalf("data source config type = %T, want types.ObjectType", dataSourceResponse.Schema.Attributes["config"].GetType())
	}
	assertModelMatchesAttributes(t, dataSourceConfig{}, dataSourceConfigType.AttrTypes)
}

func TestClusterModelsDecodeSchemaValues(t *testing.T) {
	ctx := context.Background()

	var resourceResponse resource.SchemaResponse
	NewResource().Schema(ctx, resource.SchemaRequest{}, &resourceResponse)
	if resourceResponse.Diagnostics.HasError() {
		t.Fatalf("resource schema diagnostics: %#v", resourceResponse.Diagnostics)
	}
	resourceValue := nullObjectValue(t, ctx, resourceResponse.Schema.Type())
	var resourceModel Cluster
	if diagnostics := tfsdk.ValueAs(ctx, resourceValue, &resourceModel); diagnostics.HasError() {
		t.Fatalf("decode resource model: %#v", diagnostics)
	}

	var dataSourceResponse datasource.SchemaResponse
	NewDataSource().Schema(ctx, datasource.SchemaRequest{}, &dataSourceResponse)
	if dataSourceResponse.Diagnostics.HasError() {
		t.Fatalf("data source schema diagnostics: %#v", dataSourceResponse.Diagnostics)
	}
	dataSourceValue := nullObjectValue(t, ctx, dataSourceResponse.Schema.Type())
	var dataSourceModel dataSourceCluster
	if diagnostics := tfsdk.ValueAs(ctx, dataSourceValue, &dataSourceModel); diagnostics.HasError() {
		t.Fatalf("decode data source model: %#v", diagnostics)
	}
}

func TestConfigToStatePreservesStateOnlyFields(t *testing.T) {
	ctx := context.Background()
	state := Config{
		configModel: configModel{
			Password:             types.StringValue("legacy-password"),
			NotifyKeyspaceEvents: types.StringValue(""),
		},
		PasswordWo:        types.StringNull(),
		PasswordWoVersion: types.Int64Value(2),
	}

	var diagnostics diag.Diagnostics
	state.configModel = configToState(ctx, testRedisAPIConfig(), state.configModel, &diagnostics)
	if diagnostics.HasError() {
		t.Fatalf("configToState() diagnostics: %#v", diagnostics)
	}

	if state.Password.ValueString() != "legacy-password" {
		t.Fatalf("password = %q, want legacy-password", state.Password.ValueString())
	}
	if state.NotifyKeyspaceEvents.IsNull() || state.NotifyKeyspaceEvents.ValueString() != "" {
		t.Fatalf("notify_keyspace_events = %#v, want explicit empty string", state.NotifyKeyspaceEvents)
	}
	if state.PasswordWoVersion.ValueInt64() != 2 {
		t.Fatalf("password_wo_version = %d, want 2", state.PasswordWoVersion.ValueInt64())
	}
	if !state.PasswordWo.IsNull() {
		t.Fatalf("password_wo = %#v, want null", state.PasswordWo)
	}
}

func assertModelMatchesAttributes[T any](t *testing.T, model any, attributes map[string]T) {
	t.Helper()

	modelAttributes := modelAttributeNames(t, reflect.TypeOf(model))
	if len(modelAttributes) != len(attributes) {
		t.Fatalf("%T attributes = %v, schema attributes = %v", model, modelAttributes, attributes)
	}
	for name := range attributes {
		if _, ok := modelAttributes[name]; !ok {
			t.Fatalf("%T is missing schema attribute %q", model, name)
		}
	}
}

func modelAttributeNames(t *testing.T, modelType reflect.Type) map[string]struct{} {
	t.Helper()

	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}

	result := make(map[string]struct{})
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.Anonymous {
			for name := range modelAttributeNames(t, field.Type) {
				if _, ok := result[name]; ok {
					t.Fatalf("%s has duplicate tfsdk tag %q", modelType.Name(), name)
				}
				result[name] = struct{}{}
			}
			continue
		}

		name := field.Tag.Get("tfsdk")
		if name == "" {
			t.Fatalf("%s.%s has no tfsdk tag", modelType.Name(), field.Name)
		}
		if _, ok := result[name]; ok {
			t.Fatalf("%s has duplicate tfsdk tag %q", modelType.Name(), name)
		}
		result[name] = struct{}{}
	}
	return result
}

func nullObjectValue(t *testing.T, ctx context.Context, schemaType attr.Type) types.Object {
	t.Helper()

	objectType, ok := schemaType.(types.ObjectType)
	if !ok {
		t.Fatalf("schema type = %T, want types.ObjectType", schemaType)
	}

	values := make(map[string]attr.Value, len(objectType.AttrTypes))
	for name, attributeType := range objectType.AttrTypes {
		value, err := attributeType.ValueFromTerraform(
			ctx,
			tftypes.NewValue(attributeType.TerraformType(ctx), nil),
		)
		if err != nil {
			t.Fatalf("create null value for %q: %s", name, err)
		}
		values[name] = value
	}

	result, diagnostics := types.ObjectValue(objectType.AttrTypes, values)
	if diagnostics.HasError() {
		t.Fatalf("create null schema object: %#v", diagnostics)
	}
	return result
}

func testRedisAPIConfig() *redis.ClusterConfig {
	return &redis.ClusterConfig{
		Version: "9.1-valkey",
		Redis:   &config.RedisConfigSet{UserConfig: &config.RedisConfig{}},
	}
}
