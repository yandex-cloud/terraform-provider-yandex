package yandex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func init() {
	resource.AddTestSweepers("yandex_mdb_postgresql_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_postgresql_cluster_v2",
		F:    testSweepMDBPostgreSQLClusterV2,
	})
}

func testSweepMDBPostgreSQLClusterV2(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().PostgreSQL().Cluster().List(context.Background(), &postgresql.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
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

func sweepMDBPostgreSQLClusterV2(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBPostgreSQLClusterV2Once, conf, "PostgreSQL cluster", id)
}

func sweepMDBPostgreSQLClusterV2Once(conf *Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBPostgreSQLClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	op, err := conf.sdk.MDB().PostgreSQL().Cluster().Update(ctx, &postgresql.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().PostgreSQL().Cluster().Delete(ctx, &postgresql.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
