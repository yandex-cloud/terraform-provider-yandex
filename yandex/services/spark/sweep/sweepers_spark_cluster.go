package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
	sparksdk "github.com/yandex-cloud/go-sdk/services/spark/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerSweepersSparkCluster(add func(string, *resource.Sweeper)) {
	add("yandex_spark_cluster", &resource.Sweeper{
		Name: "yandex_spark_cluster",
		F:    sweepAllSparkCluster,
	})
}

func sweepAllSparkCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := sparksdk.NewClusterClient(conf.SDK)
	resp, err := client.List(context.Background(), &spark.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sparkClusterPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Spark clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepSparkCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Spark cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

const (
	sparkClusterPageSize      = 1000
	sparkClusterDeleteTimeout = 30 * time.Minute
)

func sweepSparkCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepSparkClusterOnce, conf, "Spark cluster", id)
}

func sweepSparkClusterOnce(conf *providerconfig.Config, id string) error {
	ctxDel, cancelDel := context.WithTimeout(context.Background(), sparkClusterDeleteTimeout)
	defer cancelDel()
	client := sparksdk.NewClusterClient(conf.SDK)
	op, err := client.Delete(ctxDel, &spark.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctxDel, op, err)
}
