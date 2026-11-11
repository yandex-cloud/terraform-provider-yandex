package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	containerssdk "github.com/yandex-cloud/go-sdk/services/serverless/containers/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_serverless_containers "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/containers/pluginsdk"
)

func registerResourceYandexServerlessContainer(add func(string, *resource.Sweeper)) {
	add("yandex_serverless_container", &resource.Sweeper{
		Name: "yandex_serverless_container",
		F:    sweepAllServerlessContainer,
		Dependencies: []string{
			"yandex_iam_service_account",
		},
	})
}

func sweepAllServerlessContainer(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &containers.ListContainersRequest{FolderId: conf.FolderID}
	client := containerssdk.NewContainerClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepServerlessContainer(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Serverless Container %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepServerlessContainer(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepServerlessContainerOnce, conf, "Serverless Container", id)
}

func sweepServerlessContainerOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_serverless_containers.YandexServerlessContainerDefaultTimeout)
	defer cancel()

	client := containerssdk.NewContainerClient(conf.SDK)

	op, err := client.Delete(ctx, &containers.DeleteContainerRequest{
		ContainerId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
