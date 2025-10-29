package yandex

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var providerResources []func() resource.Resource

func GetProviderResources() []func() resource.Resource {
	return providerResources
}
