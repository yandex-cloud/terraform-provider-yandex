package providermux_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	frameworkprovider "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/providermux"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
)

func TestProductionSchema(t *testing.T) {
	pluginProvider, err := yandex.NewPluginSDKProvider()
	if err != nil {
		t.Fatal(err)
	}
	factory, err := providermux.New(context.Background(), pluginProvider, frameworkprovider.NewFrameworkProvider())
	if err != nil {
		t.Fatal(err)
	}
	response, err := factory().GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if response == nil {
		t.Fatal("provider returned no schema response")
	}
	for _, diagnostic := range response.Diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("%s: %s", diagnostic.Summary, diagnostic.Detail)
		}
	}
	for _, name := range []string{
		"yandex_dns_zone",
		"yandex_billing_cloud_binding",
		"yandex_datatransfer_transfer",
	} {
		if response.ResourceSchemas[name] == nil {
			t.Errorf("resource %q is missing from the combined schema", name)
		}
		if response.DataSourceSchemas[name] == nil {
			t.Errorf("data source %q is missing from the combined schema", name)
		}
	}
}
