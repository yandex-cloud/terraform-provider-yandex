package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	cdnsdk "github.com/yandex-cloud/go-sdk/services/cdn/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_cdn "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cdn/pluginsdk"
)

func registerResourceYandexCdnOriginGroup(add func(string, *resource.Sweeper)) {
	add("yandex_cdn_origin_group", &resource.Sweeper{
		Name: "yandex_cdn_origin_group",
		F:    sweepAllCDNOriginGroups,
		Dependencies: []string{
			"yandex_cdn_resource",
		},
	})
}

func sweepAllCDNOriginGroups(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &cdn.ListOriginGroupsRequest{FolderId: conf.FolderID}
	client := cdnsdk.NewOriginGroupClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}

	for it.Next() {
		id := it.Value().GetId()
		if !sweepCDNOriginGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep CDN Origin group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCDNOriginGroup(conf *providerconfig.Config, id int64) bool {
	return sweep.SweepWithRetryByFunc(conf, "Origin Group", func(conf *providerconfig.Config) error {
		ctx, cancel := conf.ContextWithTimeout(svc_cdn.YandexCDNOriginGroupDefaultTimeout)
		defer cancel()

		client := cdnsdk.NewOriginGroupClient(conf.SDK)

		op, err := client.Delete(ctx, &cdn.DeleteOriginGroupRequest{
			FolderId:      conf.FolderID,
			OriginGroupId: id,
		})

		return sweep.HandleSweepOperationV2(ctx, op, err)
	})
}
