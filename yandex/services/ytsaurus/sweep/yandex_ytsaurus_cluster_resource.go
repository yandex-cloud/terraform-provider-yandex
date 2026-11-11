package sweep

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ytsaurus/v1"
	ytsaurusv1sdk "github.com/yandex-cloud/go-sdk/services/ytsaurus/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerYandexYtsaurusClusterResource(add func(string, *resource.Sweeper)) {
	add("yandex_ytsaurus_cluster", &resource.Sweeper{
		Name:         "yandex_ytsaurus_cluster",
		F:            sweepAllCluster,
		Dependencies: []string{},
	})
}

func sweepAllCluster(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := ytsaurusv1sdk.NewClusterClient(conf.SDKv2).List(context.Background(), &ytsaurus.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error getting YTsaurus clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepYtsaurusCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep YTsaurus cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepYtsaurusCluster(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepYtsaurusClusterOnce, conf, "YTsaurus cluster", id)
}

func sweepYtsaurusClusterOnce(conf *provider_config.Config, id string) error {
	op, err := ytsaurusv1sdk.NewClusterClient(conf.SDKv2).Delete(context.Background(), &ytsaurus.DeleteClusterRequest{
		ClusterId: id,
	})
	_, err = op.Wait(context.Background())
	return err
}
