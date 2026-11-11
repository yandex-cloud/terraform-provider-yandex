package yandex

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestMergeRegistrationsPreservesResourcesAndSeparateNamespaces(t *testing.T) {
	providerOwned := &schema.Resource{SchemaVersion: 2}
	resource := &schema.Resource{SchemaVersion: 3}
	dataSource := &schema.Resource{}
	provider := &schema.Provider{DataSourcesMap: map[string]*schema.Resource{"yandex_client_config": providerOwned}}
	err := mergeRegistrations(provider, []serviceRegistration{{
		service:     "dns",
		resources:   map[string]*schema.Resource{"yandex_dns_zone": resource},
		dataSources: map[string]*schema.Resource{"yandex_dns_zone": dataSource},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if provider.ResourcesMap["yandex_dns_zone"] != resource || provider.DataSourcesMap["yandex_dns_zone"] != dataSource || provider.DataSourcesMap["yandex_client_config"] != providerOwned {
		t.Fatal("registration replaced a resource or a provider-owned data source")
	}
}

func TestMergeRegistrationsRejectsInvalidServiceWithoutPartialChanges(t *testing.T) {
	resource := &schema.Resource{}
	tests := []struct {
		name    string
		modules []serviceRegistration
		want    string
	}{
		{"resource collision", []serviceRegistration{{service: "first", resources: map[string]*schema.Resource{"shared": resource}}, {service: "second", resources: map[string]*schema.Resource{"shared": resource}}}, `register service "second": resource "shared" is already registered by "first"`},
		{"data source collision", []serviceRegistration{{service: "first", dataSources: map[string]*schema.Resource{"shared": resource}}, {service: "second", dataSources: map[string]*schema.Resource{"shared": resource}}}, `register service "second": data source "shared" is already registered by "first"`},
		{"provider collision", []serviceRegistration{{service: "first", dataSources: map[string]*schema.Resource{"yandex_client_config": resource}}}, `register service "first": data source "yandex_client_config" is already registered by "provider"`},
		{"nil resource", []serviceRegistration{{service: "first", resources: map[string]*schema.Resource{"missing": nil}}}, "invalid resource"},
		{"empty data source name", []serviceRegistration{{service: "first", dataSources: map[string]*schema.Resource{"": resource}}}, "invalid data source"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &schema.Provider{DataSourcesMap: map[string]*schema.Resource{"yandex_client_config": resource}}
			err := mergeRegistrations(provider, test.modules)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v, want error containing %q", err, test.want)
			}
			if provider.ResourcesMap != nil || len(provider.DataSourcesMap) != 1 || provider.DataSourcesMap["yandex_client_config"] != resource {
				t.Fatal("rejected service partially changed provider registrations")
			}
		})
	}
}
