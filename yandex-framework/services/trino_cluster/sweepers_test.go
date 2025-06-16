package trino_cluster_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	trinov1 "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	trinoClusterPageSize      = 1000
	trinoClusterDeleteTimeout = 30 * time.Minute
	trinoClusterUpdateTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_trino_cluster", &resource.Sweeper{
		Name: "yandex_trino_cluster",
		F:    testSweepMDBTrinoCluster,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepMDBTrinoCluster(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.Trino().Cluster().List(context.Background(), &trinov1.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: trinoClusterPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Trino clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBTrinoCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Trino cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBTrinoCluster(conf *config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepMDBTrinoClusterOnce, conf, "Trino cluster", id)
}

func sweepMDBTrinoClusterOnce(conf *config.Config, id string) error {
	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	ctxUpd, cancelUpd := context.WithTimeout(context.Background(), trinoClusterUpdateTimeout)
	defer cancelUpd()
	op, err := conf.SDK.Trino().Cluster().Update(ctxUpd, &trinov1.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = testhelpers.HandleSweepOperation(ctxUpd, conf, op, err)
	if err != nil && !strings.EqualFold(testhelpers.ErrorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), trinoClusterDeleteTimeout)
	defer cancelDel()
	op, err = conf.SDK.Trino().Cluster().Delete(ctxDel, &trinov1.DeleteClusterRequest{
		ClusterId: id,
	})
	return testhelpers.HandleSweepOperation(ctxDel, conf, op, err)
}
