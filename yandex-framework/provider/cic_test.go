package provider

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	testresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	yandex_gen "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cic_private_connection"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cic_public_connection"
)

func TestMain(m *testing.M) {
	testresource.TestMain(m)
}

func TestCICResourceRegistration(t *testing.T) {
	ctx := context.Background()
	original := slices.Clone(yandex_gen.GetProviderResources())
	registered := (&Provider{}).Resources(ctx)
	actual := registered[len(registered)-len(original):]
	changed := 0
	for i, factory := range original {
		if reflect.ValueOf(factory).Pointer() != reflect.ValueOf(yandex_gen.GetProviderResources()[i]).Pointer() {
			t.Fatal("generated resource registry was modified")
		}
		var metadata resource.MetadataResponse
		factory().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "yandex"}, &metadata)
		if metadata.TypeName != "yandex_cic_private_connection" {
			if reflect.ValueOf(factory).Pointer() != reflect.ValueOf(actual[i]).Pointer() {
				t.Errorf("factory changed for %s", metadata.TypeName)
			}
			continue
		}
		changed++
		if reflect.ValueOf(actual[i]).Pointer() != reflect.ValueOf(yandex_cic_private_connection.NewResourceWithMaskedLogging).Pointer() {
			t.Fatal("CIC resource is not using masked logging")
		}
		r := actual[i]()
		if _, ok := r.(resource.ResourceWithConfigure); !ok {
			t.Error("CIC resource lost Configure")
		}
		if _, ok := r.(resource.ResourceWithImportState); !ok {
			t.Error("CIC resource lost ImportState")
		}
		if _, ok := r.(resource.ResourceWithModifyPlan); !ok {
			t.Error("CIC resource lost ModifyPlan")
		}
		var before, after resource.SchemaResponse
		factory().Schema(ctx, resource.SchemaRequest{}, &before)
		r.Schema(ctx, resource.SchemaRequest{}, &after)
		if !reflect.DeepEqual(before, after) {
			t.Error("CIC resource schema changed")
		}
	}
	if changed != 1 {
		t.Fatalf("CIC resources = %d, want 1", changed)
	}
}

func TestCICDataSourceRegistration(t *testing.T) {
	ctx := context.Background()
	original := slices.Clone(yandex_gen.GetProviderDataSources())
	registered := (&Provider{}).DataSources(ctx)
	actual := registered[len(registered)-len(original):]
	expected := map[string]func() datasource.DataSource{
		"yandex_cic_private_connection": yandex_cic_private_connection.NewDataSourceWithMaskedLogging,
		"yandex_cic_public_connection":  yandex_cic_public_connection.NewDataSourceWithMaskedLogging,
	}
	changed := 0
	for i, factory := range original {
		if reflect.ValueOf(factory).Pointer() != reflect.ValueOf(yandex_gen.GetProviderDataSources()[i]).Pointer() {
			t.Fatal("generated data source registry was modified")
		}
		var metadata datasource.MetadataResponse
		factory().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "yandex"}, &metadata)
		masked, ok := expected[metadata.TypeName]
		if !ok {
			if reflect.ValueOf(factory).Pointer() != reflect.ValueOf(actual[i]).Pointer() {
				t.Errorf("factory changed for %s", metadata.TypeName)
			}
			continue
		}
		changed++
		if reflect.ValueOf(actual[i]).Pointer() != reflect.ValueOf(masked).Pointer() {
			t.Fatalf("%s is not using masked logging", metadata.TypeName)
		}
		if _, ok := actual[i]().(datasource.DataSourceWithConfigure); !ok {
			t.Errorf("%s lost Configure", metadata.TypeName)
		}
		var before, after datasource.SchemaResponse
		factory().Schema(ctx, datasource.SchemaRequest{}, &before)
		actual[i]().Schema(ctx, datasource.SchemaRequest{}, &after)
		if !reflect.DeepEqual(before, after) {
			t.Errorf("%s schema changed", metadata.TypeName)
		}
	}
	if changed != len(expected) {
		t.Fatalf("masked data sources = %d, want %d", changed, len(expected))
	}
}
