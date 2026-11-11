package sweep

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	greenplumsdk "github.com/yandex-cloud/go-sdk/services/mdb/greenplum/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	defaultMDBPageSize            = 1000
	greenplumClusterDeleteTimeout = 15 * time.Minute
)

func registerMdbGreenplumClusterV2Resource(add func(string, *resource.Sweeper)) {
	add("yandex_mdb_greenplum_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_greenplum_cluster_v2",
		F:    sweepAllMDBPostgreSQLCluster,
	})
}

func sweepAllMDBPostgreSQLCluster(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := greenplumsdk.NewClusterClient(conf.SDKv2).List(context.Background(), &greenplum.ListClustersRequest{
		FolderId:  conf.ProviderState.FolderID.ValueString(),
		PageSize:  defaultMDBPageSize,
		PageToken: "",
		Filter:    "",
	})
	if err != nil {
		return fmt.Errorf("error getting PostgreSQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !testhelpers.SweepWithRetry(sweepMDBGreenplumClusterV2, conf, "Greenplum cluster", c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep PostgreSQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBGreenplumClusterV2(conf *config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), greenplumClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	client := greenplumsdk.NewClusterClient(conf.SDKv2)
	clusterAbsent := func() bool {
		_, getErr := client.Get(ctx, &greenplum.GetClusterRequest{ClusterId: id})
		return status.Code(getErr) == codes.NotFound
	}

	op, err := client.Update(ctx, &greenplum.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = testhelpers.HandleSweepOperation(ctx, op, err)
	if err != nil && !strings.EqualFold(testhelpers.ErrorMessage(err), "no changes detected") {
		if clusterAbsent() {
			return nil
		}
		return err
	}

	deleteOp, err := client.Delete(ctx, &greenplum.DeleteClusterRequest{
		ClusterId: id,
	})
	err = testhelpers.HandleSweepOperation(ctx, deleteOp, err)
	if err != nil && clusterAbsent() {
		return nil
	}
	return err
}
