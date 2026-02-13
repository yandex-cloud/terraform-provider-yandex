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

	resp, err := conf.sdk.MDB().OpenSearch().Cluster().List(
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

	return result.ErrorOrNil()
}

func sweepMDBOpenSearchCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBOpenSearchClusterOnce, conf, "OpenSearch cluster", id)
}

func sweepMDBOpenSearchClusterOnce(conf *Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), openSearchClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.MDB().OpenSearch().Cluster().Update(ctx, &opensearch.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().OpenSearch().Cluster().Delete(ctx, &opensearch.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
