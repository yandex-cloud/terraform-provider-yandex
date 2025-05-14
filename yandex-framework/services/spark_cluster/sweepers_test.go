package spark_cluster_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
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

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepSparkCluster(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.Spark().Cluster().List(context.Background(), &spark.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
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

func sweepSparkCluster(conf *config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepSparkClusterOnce, conf, "Spark cluster", id)
}

func sweepSparkClusterOnce(conf *config.Config, id string) error {
	ctxDel, cancelDel := context.WithTimeout(context.Background(), sparkClusterDeleteTimeout)
	defer cancelDel()
	op, err := conf.SDK.Spark().Cluster().Delete(ctxDel, &spark.DeleteClusterRequest{
		ClusterId: id,
	})
	return testhelpers.HandleSweepOperation(ctxDel, conf, op, err)
}
