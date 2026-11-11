package sweep

import (
	"context"
	"fmt"
	"strings"
	"time"

	clickhousesdk "github.com/yandex-cloud/go-sdk/services/mdb/clickhouse/v1"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	yandexMDBClickHouseClusterCreateTimeout = 30 * time.Minute
	yandexMDBClickHouseClusterDeleteTimeout = 15 * time.Minute
	yandexMDBClickHouseClusterUpdateTimeout = 60 * time.Minute

	chVersionV2        = "25.8"
	chUpdatedVersionV2 = "26.3"

	sqlManagementEnabled = `
  sql_user_management     = true
  sql_database_management = true
`

	chResourceKeeperV2       = "yandex_mdb_clickhouse_cluster_v2.keeper"
	chResourceCloudStorageV2 = "yandex_mdb_clickhouse_cluster_v2.cloud"
	chResourceShardedV2      = "yandex_mdb_clickhouse_cluster_v2.bar"
	chResource               = "yandex_mdb_clickhouse_cluster_v2.foo"
	chResourceRestore        = "yandex_mdb_clickhouse_cluster_v2.restore_test"
	chResourceRemoteServers  = "yandex_mdb_clickhouse_cluster_v2.remote_servers"
	chResourceMigrateKeeper  = "yandex_mdb_clickhouse_cluster_v2.migrate_to_keeper"

	chRestoreBackupWithExtensionId = "c9qqv135s8oadbbjcq7m:c9qmqctragekko24qtt3"

	defaultMDBPageSize = 1000
)

func registerMdbClickhouseClusterV2Resource(add func(string, *resource.Sweeper)) {
	add("yandex_mdb_clickhouse_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_clickhouse_cluster_v2",
		F:    sweepAllMDBClickHouseClusterV2,
	})
}

func sweepAllMDBClickHouseClusterV2(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := clickhousesdk.NewClusterClient(conf.SDKv2).List(context.Background(), &clickhouse.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting ClickHouse clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBClickHouseClusterV2(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ClickHouse cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBClickHouseClusterV2(conf *config.Config, id string) bool {
	return test.SweepWithRetry(sweepMDBClickHouseClusterOnceV2, conf, "ClickHouse cluster", id)
}

func sweepMDBClickHouseClusterOnceV2(conf *config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBClickHouseClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	client := clickhousesdk.NewClusterClient(conf.SDKv2)
	clusterAbsent := func() bool {
		_, getErr := client.Get(ctx, &clickhouse.GetClusterRequest{ClusterId: id})
		return status.Code(getErr) == codes.NotFound
	}
	op, err := client.Update(ctx, &clickhouse.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = test.HandleSweepOperation(ctx, op, err)
	if err != nil && !strings.EqualFold(test.ErrorMessage(err), "no changes detected") {
		if clusterAbsent() {
			return nil
		}
		return err
	}

	deleteOp, err := client.Delete(ctx, &clickhouse.DeleteClusterRequest{
		ClusterId: id,
	})
	err = test.HandleSweepOperation(ctx, deleteOp, err)
	if err != nil && clusterAbsent() {
		return nil
	}
	return err
}
