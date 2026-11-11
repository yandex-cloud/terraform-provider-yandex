package framework

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cic/framework/yandex_cic_private_connection"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cic/framework/yandex_cic_public_connection"
)

func generatedResourcesWithCICLogging(ctx context.Context) []func() resource.Resource {
	resources := slices.Clone(generatedResources())
	for i, factory := range resources {
		var metadata resource.MetadataResponse
		factory().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "yandex"}, &metadata)
		if metadata.TypeName == "yandex_cic_private_connection" {
			resources[i] = yandex_cic_private_connection.NewResourceWithMaskedLogging
		}
	}
	return resources
}

func generatedDataSourcesWithCICLogging(ctx context.Context) []func() datasource.DataSource {
	dataSources := slices.Clone(generatedDataSources())
	for i, factory := range dataSources {
		var metadata datasource.MetadataResponse
		factory().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "yandex"}, &metadata)
		switch metadata.TypeName {
		case "yandex_cic_private_connection":
			dataSources[i] = yandex_cic_private_connection.NewDataSourceWithMaskedLogging
		case "yandex_cic_public_connection":
			dataSources[i] = yandex_cic_public_connection.NewDataSourceWithMaskedLogging
		}
	}
	return dataSources
}
