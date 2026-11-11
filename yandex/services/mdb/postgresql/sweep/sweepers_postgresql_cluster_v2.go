package sweep

import (
	"context"
	"fmt"
	"strings"

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

func registerSweepersPostgresqlClusterV2(add func(string, *resource.Sweeper)) {
	add("yandex_mdb_postgresql_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_postgresql_cluster_v2",
		F:    sweepAllMDBPostgreSQLClusterV2,
	})
}

func sweepAllMDBPostgreSQLClusterV2(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := postgresqlsdk.NewClusterClient(conf.SDK).List(context.Background(), &postgresql.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting PostgreSQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBPostgreSQLClusterV2(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep PostgreSQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBPostgreSQLClusterV2(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBPostgreSQLClusterV2Once, conf, "PostgreSQL cluster", id)
}

func sweepMDBPostgreSQLClusterV2Once(conf *providerconfig.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), svc_postgresql.YandexMDBPostgreSQLClusterDeleteTimeout)
	defer cancel()

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
