package sweep

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	mysqlsdk "github.com/yandex-cloud/go-sdk/services/mdb/mysql/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_mysql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/pluginsdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func registerResourceYandexMdbMysqlCluster(add func(string, *resource.Sweeper)) {
	add(mysqlResourceType, &resource.Sweeper{
		Name: mysqlResourceType,
		F:    sweepAllMDBMySQLCluster,
	})
}

func sweepAllMDBMySQLCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := mysqlsdk.NewClusterClient(conf.SDK).List(conf.Context(), &mysql.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting MySQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBMysqlCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep MySQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBMysqlCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBMysqlClusterOnce, conf, "MySQL cluster", id)
}

func sweepMDBMysqlClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_mysql.YandexMDBMySQLClusterDefaultTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	client := mysqlsdk.NewClusterClient(conf.SDK)
	clusterAbsent := func() bool {
		_, getErr := client.Get(ctx, &mysql.GetClusterRequest{ClusterId: id})
		return status.Code(getErr) == codes.NotFound
	}
	op, err := client.Update(ctx, &mysql.UpdateClusterRequest{
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

	deleteOp, err := client.Delete(ctx, &mysql.DeleteClusterRequest{
		ClusterId: id,
	})
	err = sweep.HandleSweepOperationV2(ctx, deleteOp, err)
	if err != nil && clusterAbsent() {
		return nil
	}
	return err
}

const (
	mysqlResourceType          = "yandex_mdb_mysql_cluster"
	mysqlResourceFoo           = mysqlResourceType + ".foo"
	msRestoreBackupId          = "c9qnlqr37bgp53r9pbek:mdbel77v1so5qiu199ua"
	msRestoreSourceClusterId   = "c9qnlqr37bgp53r9pbek"
	msRestoreBackupIdEncrypted = "c9qpn7idf2pvl1077h8j:mdbq3r80r4hbj8uu26sd"
)
