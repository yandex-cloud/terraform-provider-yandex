package sweep

import (
	"context"
	"fmt"
	"strings"
	"time"

	mysqlsdk "github.com/yandex-cloud/go-sdk/services/mdb/mysql/v1"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	defaultMDBPageSize                 = 1000
	msResource                         = "yandex_mdb_mysql_cluster_v2.foo"
	msSourceClusterId                  = "c9qnlqr37bgp53r9pbek"
	msRestoreBackupIdV2                = "c9qnlqr37bgp53r9pbek:mdbel77v1so5qiu199ua"
	msRestoreBackupIdEncryptedV2       = "c9qpn7idf2pvl1077h8j:mdbq3r80r4hbj8uu26sd"
	yandexMDBMySQLClusterCreateTimeout = 30 * time.Minute // TODO refactor
	yandexMDBMySQLClusterDeleteTimeout = 15 * time.Minute
	yandexMDBMySQLClusterUpdateTimeout = 60 * time.Minute
)

func registerMdbMysqlClusterV2Resource(add func(string, *resource.Sweeper)) {
	add("yandex_mdb_mysql_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_mysql_cluster_v2",
		F:    sweepAllMDBMySQLClusterV2,
	})
}

func sweepAllMDBMySQLClusterV2(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := mysqlsdk.NewClusterClient(conf.SDKv2).List(context.Background(), &mysql.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting MySQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBMySQLCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep MySQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBMySQLCluster(conf *config.Config, id string) bool {
	return test.SweepWithRetry(sweepMDBMySQLClusterOnce, conf, "MySQL cluster", id)
}

func sweepMDBMySQLClusterOnce(conf *config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBMySQLClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	client := mysqlsdk.NewClusterClient(conf.SDKv2)
	clusterAbsent := func() bool {
		_, getErr := client.Get(ctx, &mysql.GetClusterRequest{ClusterId: id})
		return status.Code(getErr) == codes.NotFound
	}
	op, err := client.Update(ctx, &mysql.UpdateClusterRequest{
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

	deleteOp, err := client.Delete(ctx, &mysql.DeleteClusterRequest{
		ClusterId: id,
	})
	err = test.HandleSweepOperation(ctx, deleteOp, err)
	if err != nil && clusterAbsent() {
		return nil
	}
	return err
}
