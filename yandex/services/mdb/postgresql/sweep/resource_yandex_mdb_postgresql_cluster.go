package sweep

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	postgresqlsdk "github.com/yandex-cloud/go-sdk/services/mdb/postgresql/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_postgresql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/postgresql/pluginsdk"
)

func registerResourceYandexMdbPostgresqlCluster(add func(string, *resource.Sweeper)) {
	add(pgResourceType, &resource.Sweeper{
		Name: pgResourceType,
		F:    sweepAllMDBPostgreSQLCluster,
	})
}

func sweepAllMDBPostgreSQLCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := postgresqlsdk.NewClusterClient(conf.SDK).List(conf.Context(), &postgresql.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting PostgreSQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBPostgreSQLCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep PostgreSQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBPostgreSQLCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBPostgreSQLClusterOnce, conf, "PostgreSQL cluster", id)
}

func sweepMDBPostgreSQLClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_postgresql.YandexMDBPostgreSQLClusterCreateTimeout)
	defer cancel()

	if err := waitPostgreSQLClusterReadyForSweep(ctx, conf, id); err != nil {
		return fmt.Errorf("cluster is not ready for sweep: %w", err)
	}

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := postgresqlsdk.NewClusterClient(conf.SDK).Update(ctx, &postgresql.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = sweep.HandleSweepOperationV2(ctx, op, err)
	if err != nil && !strings.EqualFold(sdkcommon.ErrorMessage(err), "no changes detected") {
		return err
	}

	deleteOp, err := postgresqlsdk.NewClusterClient(conf.SDK).Delete(ctx, &postgresql.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, deleteOp, err)
}

func waitPostgreSQLClusterReadyForSweep(ctx context.Context, conf *providerconfig.Config, id string) error {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for cluster %s to become ready for sweep", id)
		case <-ticker.C:
			cluster, err := postgresqlsdk.NewClusterClient(conf.SDK).Get(ctx, &postgresql.GetClusterRequest{
				ClusterId: id,
			})
			if err != nil {
				return fmt.Errorf("failed to get cluster %s state: %w", id, err)
			}

			switch cluster.Status {
			case postgresql.Cluster_CREATING,
				postgresql.Cluster_UPDATING,
				postgresql.Cluster_STOPPING,
				postgresql.Cluster_STARTING:
				continue
			default:
				return nil
			}
		}
	}
}

const (
	pgResourceType             = "yandex_mdb_postgresql_cluster"
	pgResourceFoo              = pgResourceType + ".foo"
	pgRestoreBackupId          = "c9qrbucrcvm6a50tblv2:c9q698sst87e4vhkvrsm"
	pgRestoreBackupIdEncrypted = "c9qu0h1fg6rt1jnu62ro:mdb4er3h4lqov20tedc3"
	postgresqlLatestVersion    = "18"
	postgresqlRestoreVersion   = "15"
)
