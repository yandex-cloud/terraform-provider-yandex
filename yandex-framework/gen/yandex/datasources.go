package yandex

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var providerDataSources []func() datasource.DataSource

func GetProviderDataSources() []func() datasource.DataSource {
	return providerDataSources
}
