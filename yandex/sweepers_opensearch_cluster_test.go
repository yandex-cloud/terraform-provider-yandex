package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	opensearchsdk "github.com/yandex-cloud/go-sdk/services/mdb/opensearch/v1"
)

const (
	openSearchClusterDeleteTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_mdb_opensearch_cluster", &resource.Sweeper{
		Name: "yandex_mdb_opensearch_cluster",
		F:    testSweepMDBOpenSearchCluster,
	})
}

func testSweepMDBOpenSearchCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := opensearchsdk.NewClusterClient(conf.SDK).List(
		context.Background(),
		&opensearch.ListClustersRequest{
			FolderId: conf.FolderID,
		})
	if err != nil {
		return fmt.Errorf("error getting OpenSearch clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBOpenSearchCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep OpenSearch cluster %q", c.Id))
		}
	}

	// sleep 1 minute to let the cluster be deleted and vpc update their state
	time.Sleep(1 * time.Minute)
	return result.ErrorOrNil()
}

func sweepMDBOpenSearchCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBOpenSearchClusterOnce, conf, "OpenSearch cluster", id)
}

func sweepMDBOpenSearchClusterOnce(conf *Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), openSearchClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := opensearchsdk.NewClusterClient(conf.SDK).Update(ctx, &opensearch.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperationV2(ctx, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	deleteOp, err := opensearchsdk.NewClusterClient(conf.SDK).Delete(ctx, &opensearch.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperationV2(ctx, deleteOp, err)
}
