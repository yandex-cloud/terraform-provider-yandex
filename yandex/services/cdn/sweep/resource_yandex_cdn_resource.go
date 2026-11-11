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

func registerResourceYandexCdnResource(add func(string, *resource.Sweeper)) {
	add("yandex_cdn_resource", &resource.Sweeper{
		Name: "yandex_cdn_resource",
		F:    sweepAllCDNResource,
	})
}

func sweepAllCDNResource(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &cdn.ListResourcesRequest{
		FolderId: conf.FolderID,
	}

	client := cdnsdk.NewResourceClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}

	for it.Next() {
		id := it.Value().GetId()
		if !sweepCDNResource(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep CDN resource %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCDNResource(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetryByFunc(conf, "CDN Resource", func(conf *providerconfig.Config) error {
		ctx, cancel := conf.ContextWithTimeout(svc_cdn.YandexCDNOriginGroupDefaultTimeout)
		defer cancel()

		client := cdnsdk.NewResourceClient(conf.SDK)

		op, err := client.Delete(ctx, &cdn.DeleteResourceRequest{
			ResourceId: id,
		})

		return sweep.HandleSweepOperationV2(ctx, op, err)
	})
}
