package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/framework/storage_bucket_grant"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/framework/storage_bucket_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/framework/storage_bucket_policy"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		storage_bucket_grant.NewResource,
		storage_bucket_iam_binding.NewIamBinding,
		storage_bucket_policy.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return nil
}
