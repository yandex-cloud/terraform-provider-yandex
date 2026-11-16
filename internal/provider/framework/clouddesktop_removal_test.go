package framework

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestCloudDesktopNotRegistered(t *testing.T) {
	ctx := context.Background()
	p := &Provider{}

	t.Run("resources", func(t *testing.T) {
		factories := p.Resources(ctx)
		if len(factories) == 0 {
			t.Fatal("provider resource registry is empty")
		}
		for _, factory := range factories {
			var metadata resource.MetadataResponse
			factory().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "yandex"}, &metadata)
			if strings.HasPrefix(metadata.TypeName, "yandex_cloud_desktops_") {
				t.Errorf("removed Cloud Desktop resource is still registered: %s", metadata.TypeName)
			}
		}
	})

	t.Run("data sources", func(t *testing.T) {
		factories := p.DataSources(ctx)
		if len(factories) == 0 {
			t.Fatal("provider data source registry is empty")
		}
		for _, factory := range factories {
			var metadata datasource.MetadataResponse
			factory().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "yandex"}, &metadata)
			if strings.HasPrefix(metadata.TypeName, "yandex_cloud_desktops_") {
				t.Errorf("removed Cloud Desktop data source is still registered: %s", metadata.TypeName)
			}
		}
	})
}
