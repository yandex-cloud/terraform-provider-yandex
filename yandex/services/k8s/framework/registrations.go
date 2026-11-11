package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/k8s/framework/kubernetes_marketplace_helm_release"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		kubernetes_marketplace_helm_release.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return nil
}
