package acceptance

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
)

func TestConfigRequiresConfiguredProvider(t *testing.T) {
	var nilConfig *yandex.Config
	for _, tt := range []struct {
		name      string
		meta      interface{}
		configure bool
		want      string
	}{
		{name: "not configured", want: "not configured"},
		{name: "wrong Meta type", meta: "unexpected configuration", configure: true, want: "got string"},
		{name: "typed nil", meta: nilConfig, configure: true, want: "nil *yandex.Config"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			provider := &schema.Provider{ConfigureContextFunc: func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) { return tt.meta, nil }}
			session := &Session{provider: provider}
			if tt.configure {
				if diagnostics := provider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil)); diagnostics.HasError() {
					t.Fatal(diagnostics)
				}
			}
			config, err := session.Config()
			if config != nil || err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Config() = %v, %v; want error containing %q", config, err, tt.want)
			}
		})
	}
}

func TestConfigThroughProtocol(t *testing.T) {
	session := New(t)
	expected := &yandex.Config{FolderID: "test-session-folder"}
	session.Provider().ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) { return expected, nil }
	server, err := session.Protocol6Factories()["yandex"]()
	if err != nil {
		t.Fatal(err)
	}
	response, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.Provider == nil {
		t.Fatal("provider returned no configuration schema")
	}
	for _, diagnostic := range response.Diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("%s: %s", diagnostic.Summary, diagnostic.Detail)
		}
	}
	objectType := response.Provider.ValueType().(tftypes.Object)
	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, nil)
	}
	attributes["token"] = tftypes.NewValue(tftypes.String, "t1.test.session-token")
	configuration, err := tfprotov6.NewDynamicValue(objectType, tftypes.NewValue(objectType, attributes))
	if err != nil {
		t.Fatal(err)
	}
	configured, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: &configuration, TerraformVersion: "1.12.0"})
	if err != nil {
		t.Fatal(err)
	}
	if configured == nil {
		t.Fatal("provider returned no configuration response")
	}
	for _, diagnostic := range configured.Diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("%s: %s", diagnostic.Summary, diagnostic.Detail)
		}
	}
	actual, err := session.Config()
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatal("Session.Config returned a different configuration than ConfigureProvider installed")
	}
}

func TestSessionsHaveIndependentProviders(t *testing.T) {
	first, second := New(t), New(t)
	if first.Provider() == second.Provider() {
		t.Fatal("sessions share a provider")
	}
	delete(first.Provider().ResourcesMap, "yandex_dns_zone")
	delete(first.Provider().DataSourcesMap, "yandex_dns_zone")
	if second.Provider().ResourcesMap["yandex_dns_zone"] == nil || second.Provider().DataSourcesMap["yandex_dns_zone"] == nil {
		t.Fatal("sessions share registration maps")
	}
	if second.Provider().Meta() != nil {
		t.Fatal("new session has provider configuration")
	}
}
