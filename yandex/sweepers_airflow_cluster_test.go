package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	afv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/airflow/v1"
)

const (
	airflowClusterPageSize      = 1000
	airflowClusterDeleteTimeout = 30 * time.Minute
	airflowClusterUpdateTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_airflow_cluster", &resource.Sweeper{
		Name: "yandex_airflow_cluster",
		F:    testSweepMDBAirflowCluster,
	})
}

func testSweepMDBAirflowCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.Airflow().Cluster().List(context.Background(), &afv1.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: airflowClusterPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Airflow clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBAirflowCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Airflow cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBAirflowCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBAirflowClusterOnce, conf, "Airflow cluster", id)
}

func sweepMDBAirflowClusterOnce(conf *Config, id string) error {
	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	ctxUpd, cancelUpd := context.WithTimeout(context.Background(), airflowClusterUpdateTimeout)
	defer cancelUpd()
	op, err := conf.sdk.Airflow().Cluster().Update(ctxUpd, &afv1.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctxUpd, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), airflowClusterDeleteTimeout)
	defer cancelDel()
	op, err = conf.sdk.Airflow().Cluster().Delete(ctxDel, &afv1.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctxDel, conf, op, err)
}
