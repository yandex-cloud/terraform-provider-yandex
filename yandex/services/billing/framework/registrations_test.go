package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestFrameworkBindingRetainsCloudParameters(t *testing.T) {
	ctx := context.Background()
	res := Resources()[0]()
	var metadata resource.MetadataResponse
	res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "yandex"}, &metadata)
	if metadata.TypeName != "yandex_billing_cloud_binding" {
		t.Fatalf("unexpected resource name %q", metadata.TypeName)
	}
	var schema resource.SchemaResponse
	res.Schema(ctx, resource.SchemaRequest{}, &schema)
	cloudID, ok := schema.Schema.Attributes["cloud_id"]
	if schema.Diagnostics.HasError() || !ok || !cloudID.IsRequired() {
		t.Fatalf("resource lost the required cloud_id binding parameter: %v", schema.Diagnostics)
	}
	ds := DataSources()[0]()
	var dsMetadata datasource.MetadataResponse
	ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "yandex"}, &dsMetadata)
	if dsMetadata.TypeName != "yandex_billing_cloud_binding" {
		t.Fatalf("unexpected data source name %q", dsMetadata.TypeName)
	}
	var dsSchema datasource.SchemaResponse
	ds.Schema(ctx, datasource.SchemaRequest{}, &dsSchema)
	if dsSchema.Diagnostics.HasError() || dsSchema.Schema.Attributes["cloud_id"] == nil {
		t.Fatalf("data source lost cloud_id binding parameter: %v", dsSchema.Diagnostics)
	}
}
