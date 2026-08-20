package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	spqrsdk "github.com/yandex-cloud/go-sdk/services/mdb/spqr/v1"
)

const (
	yandexMDBShardedPostgreSQLClusterDeleteTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_mdb_sharded_postgresql_cluster", &resource.Sweeper{
		Name: "yandex_mdb_sharded_postgresql_cluster",
		F:    testSweepMDBShardedPostgreSQLCluster,
	})
}

func testSweepMDBShardedPostgreSQLCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := spqrsdk.NewClusterClient(conf.SDK).List(context.Background(), &spqr.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Sharded PostgreSQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBShardedPostgreSQLCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Sharded PostgreSQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBShardedPostgreSQLCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBShardedPostgreSQLClusterOnce, conf, "Sharded PostgreSQL cluster", id)
}

func sweepMDBShardedPostgreSQLClusterOnce(conf *Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBShardedPostgreSQLClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	op, err := spqrsdk.NewClusterClient(conf.SDK).Update(ctx, &spqr.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperationV2(ctx, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	deleteOp, err := spqrsdk.NewClusterClient(conf.SDK).Delete(ctx, &spqr.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperationV2(ctx, deleteOp, err)
}
