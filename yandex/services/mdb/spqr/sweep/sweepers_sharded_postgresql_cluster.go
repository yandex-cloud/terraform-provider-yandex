package sweep

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
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerSweepersShardedPostgresqlCluster(add func(string, *resource.Sweeper)) {
	add("yandex_mdb_sharded_postgresql_cluster", &resource.Sweeper{
		Name: "yandex_mdb_sharded_postgresql_cluster",
		F:    sweepAllMDBShardedPostgreSQLCluster,
	})
}

func sweepAllMDBShardedPostgreSQLCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := spqrsdk.NewClusterClient(conf.SDK).List(context.Background(), &spqr.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
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

func sweepMDBShardedPostgreSQLCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBShardedPostgreSQLClusterOnce, conf, "Sharded PostgreSQL cluster", id)
}

func sweepMDBShardedPostgreSQLClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBShardedPostgreSQLClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	op, err := spqrsdk.NewClusterClient(conf.SDK).Update(ctx, &spqr.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = sweep.HandleSweepOperationV2(ctx, op, err)
	if err != nil && !strings.EqualFold(sdkcommon.ErrorMessage(err), "no changes detected") {
		return err
	}

	deleteOp, err := spqrsdk.NewClusterClient(conf.SDK).Delete(ctx, &spqr.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, deleteOp, err)
}

const (
	yandexMDBShardedPostgreSQLClusterDeleteTimeout = 30 * time.Minute
)
