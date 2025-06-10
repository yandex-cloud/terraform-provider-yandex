package metastore_cluster_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	msv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/metastore/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	metastoreClusterPageSize      = 1000
	metastoreClusterDeleteTimeout = 30 * time.Minute
	metastoreClusterUpdateTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_metastore_cluster", &resource.Sweeper{
		Name: "yandex_metastore_cluster",
		F:    testSweepMDBMetastoreCluster,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepMDBMetastoreCluster(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.Metastore().Cluster().List(context.Background(), &msv1.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: metastoreClusterPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Metastore clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBMetastoreCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Metastore cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBMetastoreCluster(conf *config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepMDBMetastoreClusterOnce, conf, "Metastore cluster", id)
}

func sweepMDBMetastoreClusterOnce(conf *config.Config, id string) error {
	ctxUpd, cancelUpd := context.WithTimeout(context.Background(), metastoreClusterUpdateTimeout)
	defer cancelUpd()
	op, err := conf.SDK.Metastore().Cluster().Update(ctxUpd, &msv1.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &field_mask.FieldMask{Paths: []string{"deletion_protection"}},
	})
	err = testhelpers.HandleSweepOperation(ctxUpd, conf, op, err)
	if err != nil && !strings.EqualFold(testhelpers.ErrorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), metastoreClusterDeleteTimeout)
	defer cancelDel()
	op, err = conf.SDK.Metastore().Cluster().Delete(ctxDel, &msv1.DeleteClusterRequest{
		ClusterId: id,
	})
	return testhelpers.HandleSweepOperation(ctxDel, conf, op, err)
}
