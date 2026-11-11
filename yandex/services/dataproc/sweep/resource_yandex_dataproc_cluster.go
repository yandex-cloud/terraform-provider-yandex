package sweep

import (
	"fmt"
	"strings"

	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
	dataprocsdk "github.com/yandex-cloud/go-sdk/services/dataproc/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_data_proc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/dataproc/pluginsdk"
)

func registerResourceYandexDataprocCluster(add func(string, *resource.Sweeper)) {
	add(dataprocResourceType, &resource.Sweeper{
		Name: dataprocResourceType,
		F:    sweepAllDataprocCluster,
	})
}

func sweepAllDataprocCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	client := dataprocsdk.NewClusterClient(conf.SDK)

	resp, err := client.List(conf.Context(), &dataproc.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Yandex Data Processing clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepDataprocCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Yandex Data Processing cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepDataprocCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepDataprocClusterOnce, conf, "Yandex Data Processing cluster", id)
}

func sweepDataprocClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_data_proc.YandexDataprocClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	client := dataprocsdk.NewClusterClient(conf.SDK)

	updateOp, err := client.Update(ctx, &dataproc.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = sweep.HandleSweepOperationV2(ctx, updateOp, err)
	if err != nil && !strings.EqualFold(sdkcommon.ErrorMessage(err), "no changes detected") {
		return err
	}

	deleteOp, err := client.Delete(ctx, &dataproc.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, deleteOp, err)
}

var testDataprocZone = "ru-central1-b"

const dataprocResourceType = "yandex_dataproc_cluster"
