package mdb_mysql_cluster_beta

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

const defaultMDBPageSize = 1000

// ==============================================================================
//                                 CLUSTER
// ==============================================================================

func readCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) *mysql.Cluster {
	cluster, err := sdk.MDB().MySQL().Cluster().Get(ctx, &mysql.GetClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get MySQL cluster:"+err.Error(),
		)
		return nil
	}
	return cluster
}

func createCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, request *mysql.CreateClusterRequest) string {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MySQL().Cluster().Create(ctx, request)
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create MySQL cluster: "+err.Error(),
		)
		return ""
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create MySQL cluster: "+err.Error(),
		)
		return ""
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Failed to retrieve operation metadata: "+err.Error(),
		)
		return ""
	}

	md, ok := protoMetadata.(*mysql.CreateClusterMetadata)
	if !ok {
		diag.AddError(
			"Failed to Create resource",
			"Failed to retrieve cluster_id",
		)
		return ""
	}

	return md.ClusterId
}

func updateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, request *mysql.UpdateClusterRequest) {
	if request == nil || request.UpdateMask == nil || len(request.UpdateMask.Paths) == 0 {
		return
	}

	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MySQL().Cluster().Update(ctx, request)
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update MySQL cluster: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update MySQL cluster: "+err.Error(),
		)
		return
	}
}

func deleteCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MySQL().Cluster().Delete(ctx, &mysql.DeleteClusterRequest{
			ClusterId: cid,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete MySQL cluster:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete MySQL cluster:"+err.Error(),
		)
	}
}

// ==============================================================================
//                                     HOST
// ==============================================================================

func retryListMySQLHostsInner(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, attempt int, maxAttempt int, condition func([]*mysql.Host) bool) ([]*mysql.Host, error) {
	log.Printf("[DEBUG] Try ListMySQLHosts, attempt: %d", attempt)
	hosts, err := listHosts(ctx, sdk, diag, cid)
	if condition(hosts) || maxAttempt <= attempt {
		return hosts, err // We tried to do our best
	}

	timeout := int(math.Pow(2, float64(attempt)))
	log.Printf("[DEBUG] Condition failed, waiting %ds before the next attempt", timeout)
	time.Sleep(time.Second * time.Duration(timeout))

	return retryListMySQLHostsInner(ctx, sdk, diag, cid, attempt+1, maxAttempt, condition)
}

// retry with 1, 2, 4, 8, 16, 32, 64, 128 seconds if no succeess
// while at least one host is unknown and there is no master
func RetryListMySQLHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) ([]*mysql.Host, error) {
	attempts := 7
	return retryListMySQLHostsInner(ctx, sdk, diag, cid, 0, attempts, func(hosts []*mysql.Host) bool {
		masterExists := false
		for _, host := range hosts {
			// Check that every host has a role
			if host.Role == mysql.Host_ROLE_UNKNOWN {
				return false
			}
			// And one of them is master
			if host.Role == mysql.Host_MASTER {
				masterExists = true
			}
		}
		return masterExists
	})
}

// Do not use. Use RetryListMySQLHosts instead
func listHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) ([]*mysql.Host, error) {
	hosts := []*mysql.Host{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().MySQL().Cluster().ListHosts(ctx, &mysql.ListClusterHostsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diag.AddError(
				"Failed to List MySQL Hosts",
				"Error while requesting API to get MySQL host:"+err.Error(),
			)
			return nil, err
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts, nil
}

func addHost(ctx context.Context, sdk *ycsdk.SDK, cid string, hostSpec *mysql.HostSpec) (*mysql.AddClusterHostsMetadata, diag.Diagnostics) {
	var diag diag.Diagnostics
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MySQL().Cluster().AddHosts(ctx, &mysql.AddClusterHostsRequest{
			ClusterId: cid,
			HostSpecs: []*mysql.HostSpec{hostSpec},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to create MySQL host:"+err.Error(),
		)
		return nil, diag
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to create MySQL host:"+err.Error(),
		)
		return nil, diag
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while get MySQL host create operation metadata"+err.Error(),
		)
		return nil, diag
	}

	md, ok := protoMetadata.(*mysql.AddClusterHostsMetadata)
	if !ok {
		diag.AddError(
			"Failed to Update resource",
			"Error while cast MySQL host create operation metadata. Expected *mysql.CreateHostMetadata, got "+fmt.Sprintf("%T", protoMetadata),
		)
		return nil, diag
	}
	return md, diag
}

func updateHost(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, spec *mysql.UpdateHostSpec) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MySQL().Cluster().UpdateHosts(ctx, &mysql.UpdateClusterHostsRequest{
			ClusterId:       cid,
			UpdateHostSpecs: []*mysql.UpdateHostSpec{spec},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to create MySQL host:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create MySQL host:"+err.Error(),
		)
		return
	}
}

func deleteHost(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, hostname string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().MySQL().Cluster().DeleteHosts(ctx, &mysql.DeleteClusterHostsRequest{
			ClusterId: cid,
			HostNames: []string{hostname},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to update resource",
			"Error while requesting API to delete MySQL host:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete MySQL host:"+err.Error(),
		)
	}
}
