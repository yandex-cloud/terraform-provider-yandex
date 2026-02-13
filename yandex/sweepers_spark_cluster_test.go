package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
)

const (
	sparkClusterPageSize      = 1000
	sparkClusterDeleteTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_spark_cluster", &resource.Sweeper{
		Name: "yandex_spark_cluster",
		F:    testSweepSparkCluster,
	})
}

func testSweepSparkCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.Spark().Cluster().List(context.Background(), &spark.ListClustersRequest{
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

func sweepSparkCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepSparkClusterOnce, conf, "Spark cluster", id)
}

func sweepSparkClusterOnce(conf *Config, id string) error {
	ctxDel, cancelDel := context.WithTimeout(context.Background(), sparkClusterDeleteTimeout)
	defer cancelDel()
	op, err := conf.sdk.Spark().Cluster().Delete(ctxDel, &spark.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctxDel, conf, op, err)
}
