package airflow_cluster_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	afv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/airflow/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	airflowClusterPageSize      = 1000
	airflowClusterDeleteTimeout = 30 * time.Minute
	airflowClusterUpdateTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_airflow_cluster", &resource.Sweeper{
		Name: "yandex_airflow_cluster",
		F:    testSweepMDBAirflowCluster,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepMDBAirflowCluster(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.Airflow().Cluster().List(context.Background(), &afv1.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: airflowClusterPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Airflow clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBAirflowCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Airflow cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBAirflowCluster(conf *config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepMDBAirflowClusterOnce, conf, "Airflow cluster", id)
}

func sweepMDBAirflowClusterOnce(conf *config.Config, id string) error {
	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	ctxUpd, cancelUpd := context.WithTimeout(context.Background(), airflowClusterUpdateTimeout)
	defer cancelUpd()
	op, err := conf.SDK.Airflow().Cluster().Update(ctxUpd, &afv1.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = testhelpers.HandleSweepOperation(ctxUpd, conf, op, err)
	if err != nil && !strings.EqualFold(testhelpers.ErrorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), airflowClusterDeleteTimeout)
	defer cancelDel()
	op, err = conf.SDK.Airflow().Cluster().Delete(ctxDel, &afv1.DeleteClusterRequest{
		ClusterId: id,
	})
	return testhelpers.HandleSweepOperation(ctxDel, conf, op, err)
}
