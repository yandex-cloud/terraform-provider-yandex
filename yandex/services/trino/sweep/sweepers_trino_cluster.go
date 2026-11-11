package sweep

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	trinov1 "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	trinosdk "github.com/yandex-cloud/go-sdk/services/trino/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerSweepersTrinoCluster(add func(string, *resource.Sweeper)) {
	add("yandex_trino_cluster", &resource.Sweeper{
		Name: "yandex_trino_cluster",
		F:    sweepAllMDBTrinoCluster,
	})
}

func sweepAllMDBTrinoCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := trinosdk.NewClusterClient(conf.SDK)
	resp, err := client.List(context.Background(), &trinov1.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: trinoClusterPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Trino clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBTrinoCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Trino cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBTrinoCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBTrinoClusterOnce, conf, "Trino cluster", id)
}

func sweepMDBTrinoClusterOnce(conf *providerconfig.Config, id string) error {
	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	ctxUpd, cancelUpd := context.WithTimeout(context.Background(), trinoClusterUpdateTimeout)
	defer cancelUpd()
	client := trinosdk.NewClusterClient(conf.SDK)
	updateOp, err := client.Update(ctxUpd, &trinov1.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = sweep.HandleSweepOperationV2(ctxUpd, updateOp, err)
	if err != nil && !strings.EqualFold(sdkcommon.ErrorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), trinoClusterDeleteTimeout)
	defer cancelDel()
	deleteOp, err := client.Delete(ctxDel, &trinov1.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctxDel, deleteOp, err)
}

const (
	trinoClusterPageSize      = 1000
	trinoClusterDeleteTimeout = 30 * time.Minute
	trinoClusterUpdateTimeout = 30 * time.Minute
)
