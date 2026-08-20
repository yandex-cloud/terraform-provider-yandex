package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	msv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/metastore/v1"
	metastoresdk "github.com/yandex-cloud/go-sdk/services/metastore/v1"
)

const (
	metastoreClusterPageSize      = 1000
	metastoreClusterDeleteTimeout = 30 * time.Minute
	metastoreClusterUpdateTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_metastore_cluster", &resource.Sweeper{
		Name: "yandex_metastore_cluster",
		F:    testSweepMDBMetastoreCluster,
	})
}

func testSweepMDBMetastoreCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := metastoresdk.NewClusterClient(conf.SDK)
	resp, err := client.List(context.Background(), &msv1.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: metastoreClusterPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Metastore clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBMetastoreCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Metastore cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBMetastoreCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBMetastoreClusterOnce, conf, "Metastore cluster", id)
}

func sweepMDBMetastoreClusterOnce(conf *Config, id string) error {
	ctxUpd, cancelUpd := context.WithTimeout(context.Background(), metastoreClusterUpdateTimeout)
	defer cancelUpd()
	client := metastoresdk.NewClusterClient(conf.SDK)
	updateOp, err := client.Update(ctxUpd, &msv1.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &field_mask.FieldMask{Paths: []string{"deletion_protection"}},
	})
	err = handleSweepOperationV2(ctxUpd, updateOp, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), metastoreClusterDeleteTimeout)
	defer cancelDel()
	deleteOp, err := client.Delete(ctxDel, &msv1.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperationV2(ctxDel, deleteOp, err)
}
