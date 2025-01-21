package mdb_postgresql_cluster_beta

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

const defaultMDBPageSize = 1000

// ==============================================================================
//                                 CLUSTER
// ==============================================================================

func readCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) *postgresql.Cluster {
	cluster, err := sdk.MDB().PostgreSQL().Cluster().Get(ctx, &postgresql.GetClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get PostgreSQL cluster:"+err.Error(),
		)
		return nil
	}
	return cluster
}

func createCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, request *postgresql.CreateClusterRequest) string {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().PostgreSQL().Cluster().Create(ctx, request)
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create PostgreSQL cluster: "+err.Error(),
		)
		return ""
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create PostgreSQL cluster: "+err.Error(),
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

	md, ok := protoMetadata.(*postgresql.CreateClusterMetadata)
	if !ok {
		diag.AddError(
			"Failed to Create resource",
			"Failed to retrieve cluster_id",
		)
		return ""
	}

	return md.ClusterId
}

func updateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, request *postgresql.UpdateClusterRequest) {
	if request == nil || request.UpdateMask == nil || len(request.UpdateMask.Paths) == 0 {
		return
	}

	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().PostgreSQL().Cluster().Update(ctx, request)
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update PostgreSQL cluster: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update PostgreSQL cluster: "+err.Error(),
		)
		return
	}
}

func deleteCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().PostgreSQL().Cluster().Delete(ctx, &postgresql.DeleteClusterRequest{
			ClusterId: cid,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete PostgreSQL cluster:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete PostgreSQL cluster:"+err.Error(),
		)
	}
}

// ==============================================================================
//                                     HOST
// ==============================================================================

func retryListPGHostsInner(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, attempt int, maxAttempt int, condition func([]*postgresql.Host) bool) ([]*postgresql.Host, error) {
	log.Printf("[DEBUG] Try ListPGHosts, attempt: %d", attempt)
	hosts, err := listHosts(ctx, sdk, diag, cid)
	if condition(hosts) || maxAttempt <= attempt {
		return hosts, err // We tried to do our best
	}

	timeout := int(math.Pow(2, float64(attempt)))
	log.Printf("[DEBUG] Condition failed, waiting %ds before the next attempt", timeout)
	time.Sleep(time.Second * time.Duration(timeout))

	return retryListPGHostsInner(ctx, sdk, diag, cid, attempt+1, maxAttempt, condition)
}

// retry with 1, 2, 4, 8, 16, 32, 64, 128 seconds if no succeess
// while at least one host is unknown and there is no master
func RetryListPGHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) ([]*postgresql.Host, error) {
	attempts := 7
	return retryListPGHostsInner(ctx, sdk, diag, cid, 0, attempts, func(hosts []*postgresql.Host) bool {
		masterExists := false
		for _, host := range hosts {
			// Check that every host has a role
			if host.Role == postgresql.Host_ROLE_UNKNOWN {
				return false
			}
			// And one of them is master
			if host.Role == postgresql.Host_MASTER {
				masterExists = true
			}
		}
		return masterExists
	})
}

// Do not use. Use RetryListPGHosts instead
func listHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) ([]*postgresql.Host, error) {
	hosts := []*postgresql.Host{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().PostgreSQL().Cluster().ListHosts(ctx, &postgresql.ListClusterHostsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diag.AddError(
				"Failed to List PostgreSQL Hosts",
				"Error while requesting API to get PostgreSQL host:"+err.Error(),
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

func addHost(ctx context.Context, sdk *ycsdk.SDK, cid string, hostSpec *postgresql.HostSpec) (*postgresql.AddClusterHostsMetadata, diag.Diagnostics) {
	var diag diag.Diagnostics
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().PostgreSQL().Cluster().AddHosts(ctx, &postgresql.AddClusterHostsRequest{
			ClusterId: cid,
			HostSpecs: []*postgresql.HostSpec{hostSpec},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to create PostgreSQL host:"+err.Error(),
		)
		return nil, diag
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to create PostgreSQL host:"+err.Error(),
		)
		return nil, diag
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while get PostgreSQL host create operation metadata"+err.Error(),
		)
		return nil, diag
	}

	md, ok := protoMetadata.(*postgresql.AddClusterHostsMetadata)
	if !ok {
		diag.AddError(
			"Failed to Update resource",
			"Error while cast PostgreSQL host create operation metadata. Expected *postgresql.CreateHostMetadata, got "+fmt.Sprintf("%T", protoMetadata),
		)
		return nil, diag
	}
	return md, diag
}

func updateHost(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, spec *postgresql.UpdateHostSpec) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().PostgreSQL().Cluster().UpdateHosts(ctx, &postgresql.UpdateClusterHostsRequest{
			ClusterId:       cid,
			UpdateHostSpecs: []*postgresql.UpdateHostSpec{spec},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to create PostgreSQL host:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create PostgreSQL host:"+err.Error(),
		)
		return
	}
}

func deleteHost(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, hostname string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().PostgreSQL().Cluster().DeleteHosts(ctx, &postgresql.DeleteClusterHostsRequest{
			ClusterId: cid,
			HostNames: []string{hostname},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to update resource",
			"Error while requesting API to delete PostgreSQL host:"+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete PostgreSQL host:"+err.Error(),
		)
	}
}
