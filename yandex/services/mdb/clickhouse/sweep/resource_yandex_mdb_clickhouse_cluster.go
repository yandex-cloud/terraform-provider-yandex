package sweep

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhousesdk "github.com/yandex-cloud/go-sdk/services/mdb/clickhouse/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_clickhouse "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/pluginsdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func registerResourceYandexMdbClickhouseCluster(add func(string, *resource.Sweeper)) {
	add(chResourceType, &resource.Sweeper{
		Name: chResourceType,
		F:    sweepAllMDBClickHouseCluster,
	})
}

func sweepAllMDBClickHouseCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := clickhousesdk.NewClusterClient(conf.SDK).List(conf.Context(), &clickhouse.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting ClickHouse clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBClickHouseCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ClickHouse cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBClickHouseCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBClickHouseClusterOnce, conf, "ClickHouse cluster", id)
}

func sweepMDBClickHouseClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_clickhouse.YandexMDBClickHouseClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	client := clickhousesdk.NewClusterClient(conf.SDK)
	clusterAbsent := func() bool {
		_, getErr := client.Get(ctx, &clickhouse.GetClusterRequest{ClusterId: id})
		return status.Code(getErr) == codes.NotFound
	}
	op, err := client.Update(ctx, &clickhouse.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = sweep.HandleSweepOperationV2(ctx, op, err)
	if err != nil && !strings.EqualFold(sdkcommon.ErrorMessage(err), "no changes detected") {
		if clusterAbsent() {
			return nil
		}
		return err
	}

	deleteOp, err := client.Delete(ctx, &clickhouse.DeleteClusterRequest{
		ClusterId: id,
	})
	err = sweep.HandleSweepOperationV2(ctx, deleteOp, err)
	if err != nil && clusterAbsent() {
		return nil
	}
	return err
}

const (
	chVersion        = "25.8"
	chUpdatedVersion = "26.3"

	chResourceType         = "yandex_mdb_clickhouse_cluster"
	chResourceFoo          = chResourceType + ".foo"
	chResourceSharded      = chResourceType + ".bar"
	chResourceCloudStorage = chResourceType + ".cloud"
	chResourceKeeper       = chResourceType + ".keeper"

	MaintenanceWindowAnytime = "type = \"ANYTIME\""
	MaintenanceWindowWeekly  = "type = \"WEEKLY\"\n    day  = \"FRI\"\n    hour = 20"
)
