package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/framework/vpc_security_group_rule"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		vpc_security_group_rule.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		vpc_security_group_rule.NewDataSource,
	}
}
