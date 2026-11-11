package sweep

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	greenplumsdk "github.com/yandex-cloud/go-sdk/services/mdb/greenplum/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	iamsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/sweep"
	svc_greenplum "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/greenplum/pluginsdk"
)

func registerResourceYandexMdbGreenplumCluster(add func(string, *resource.Sweeper)) {
	add(greenplumResourceType, &resource.Sweeper{
		Name: greenplumResourceType,
		F:    sweepAllMDBGreenplumCluster,
	})
}

func sweepAllMDBGreenplumCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := greenplumsdk.NewClusterClient(conf.SDK).List(conf.Context(), &greenplum.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Greenplum clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBGreenplumCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Greenplum cluster %q", c.Id))
		}
	}

	iamClient := iamsdk.NewServiceAccountClient(conf.SDK)

	sas, err := iamClient.Iterator(conf.Context(), &iam.ListServiceAccountsRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	}).TakeAll()
	if err != nil {
		return fmt.Errorf("error getting service account: %s", err)
	}

	for _, sa := range sas {
		if strings.Contains(strings.ToLower(sa.Name), strings.ToLower(greenplumSANamePrefix)) {
			if !iamsweep.SweepIAMServiceAccount(conf, sa.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep service account %q", sa.Id))
			}
		}
	}

	return result.ErrorOrNil()
}

const (
	greenplumResourceType = "yandex_mdb_greenplum_cluster"
	greenplumResourceFoo  = greenplumResourceType + ".foo"
	greenplumSANamePrefix = "sa-for-"
)

func sweepMDBGreenplumCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBGreenplumClusterOnce, conf, "Greenplum cluster", id)
}

func sweepMDBGreenplumClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_greenplum.YandexMDBGreenplumClusterDefaultTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	client := greenplumsdk.NewClusterClient(conf.SDK)
	clusterAbsent := func() bool {
		_, getErr := client.Get(ctx, &greenplum.GetClusterRequest{ClusterId: id})
		return status.Code(getErr) == codes.NotFound
	}
	op, err := client.Update(ctx, &greenplum.UpdateClusterRequest{
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

	deleteOp, err := client.Delete(ctx, &greenplum.DeleteClusterRequest{
		ClusterId: id,
	})
	err = sweep.HandleSweepOperationV2(ctx, deleteOp, err)
	if err != nil && clusterAbsent() {
		return nil
	}
	return err
}
