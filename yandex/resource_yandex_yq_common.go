package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

func flattenYandexYQCommonMeta(
	d *schema.ResourceData,
	meta *Ydb_FederatedQuery.CommonMeta,
) error {
	d.SetId(meta.GetId())

	return nil
}
