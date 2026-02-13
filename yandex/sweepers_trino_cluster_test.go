package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	trinov1 "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
)

const (
	trinoClusterPageSize      = 1000
	trinoClusterDeleteTimeout = 30 * time.Minute
	trinoClusterUpdateTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_trino_cluster", &resource.Sweeper{
		Name: "yandex_trino_cluster",
		F:    testSweepMDBTrinoCluster,
	})
}

func testSweepMDBTrinoCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.Trino().Cluster().List(context.Background(), &trinov1.ListClustersRequest{
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

func sweepMDBTrinoCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBTrinoClusterOnce, conf, "Trino cluster", id)
}

func sweepMDBTrinoClusterOnce(conf *Config, id string) error {
	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	ctxUpd, cancelUpd := context.WithTimeout(context.Background(), trinoClusterUpdateTimeout)
	defer cancelUpd()
	op, err := conf.sdk.Trino().Cluster().Update(ctxUpd, &trinov1.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctxUpd, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), trinoClusterDeleteTimeout)
	defer cancelDel()
	op, err = conf.sdk.Trino().Cluster().Delete(ctxDel, &trinov1.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctxDel, conf, op, err)
}
