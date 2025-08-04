package mdb_sharded_postgresql_cluster

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

var shardedPostgreSQLAPI = ShardedPostgreSQLAPI{}

type ShardedPostgreSQLAPI struct{}

const defaultMDBPageSize = 1000

// ==============================================================================
//                                 CLUSTER
// ==============================================================================

func (r *ShardedPostgreSQLAPI) GetCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) *spqr.Cluster {
	db, err := sdk.MDB().SPQR().Cluster().Get(ctx, &spqr.GetClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diags.AddError(
			"Failed to read resource",
			fmt.Sprintf("Error while requesting API to read ShardedPostgresql cluster %q: %s", cid, err.Error()),
		)
		return nil
	}
	return db
}

func (p *ShardedPostgreSQLAPI) CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *spqr.CreateClusterRequest) string {
	op, err := sdk.WrapOperation(sdk.MDB().SPQR().Cluster().Create(ctx, req))
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ShardedPostgresql cluster: %s", err.Error()),
		)
		return ""
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata: %s", op.Id(), err.Error()),
		)
		return ""
	}

	md, ok := protoMetadata.(*spqr.CreateClusterMetadata)
	if !ok {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata", op.Id()),
		)
		return ""
	}

	log.Printf("[DEBUG] Creating Sharded Postgresql Cluster %q", md.ClusterId)

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ShardedPostgresql cluster: %s", op.Id(), err.Error()),
		)
		return ""
	}

	return md.ClusterId
}

func (p *ShardedPostgreSQLAPI) UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *spqr.UpdateClusterRequest) {
	if req == nil || len(req.UpdateMask.Paths) == 0 {
		return
	}

	op, err := sdk.WrapOperation(sdk.MDB().SPQR().Cluster().Update(ctx, req))
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ShardedPostgresql cluster: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ShardedPostgresql cluster: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (p *ShardedPostgreSQLAPI) DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) {
	op, err := sdk.WrapOperation(sdk.MDB().SPQR().Cluster().Delete(ctx, &spqr.DeleteClusterRequest{
		ClusterId: cid,
	}))

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ShardedPostgresql cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ShardedPostgresql cluster %q: %s", op.Id(), cid, err.Error()),
		)
	}
}

func (p *ShardedPostgreSQLAPI) CreateHostsWithSubclusterCheck(
	ctx context.Context,
	sdk *ycsdk.SDK,
	diag *diag.Diagnostics,
	cid string, specs []*spqr.HostSpec,
	resources map[spqr.Host_Type]*spqr.Resources,
) {
	hosts, err := sdk.MDB().SPQR().Cluster().ListHosts(ctx, &spqr.ListClusterHostsRequest{ClusterId: cid})
	if err != nil {
		diag.AddError(
			"Failed to list cluster hosts",
			fmt.Sprintf("Error while requesting API to list hosts in ShardedPostgresql cluster %q: %s", cid, err.Error()),
		)
		return
	}

	currentHosts := make(map[spqr.Host_Type][]*spqr.Host)
	for _, host := range hosts.Hosts {
		currentHosts[host.Type] = append(currentHosts[host.Type], host)
	}

	newHosts := make(map[spqr.Host_Type][]*spqr.HostSpec)
	for _, spec := range specs {
		newHosts[spec.Type] = append(newHosts[spec.Type], spec)
	}

	addClusterHosts := make([]*spqr.HostSpec, 0, len(specs))
	// Create new subclusters
	for t, specs := range newHosts {
		if _, ok := currentHosts[t]; ok {
			addClusterHosts = append(addClusterHosts, specs...)
		} else {
			res, ok := resources[t]
			if !ok {
				diag.AddError(
					"Failed to create hosts",
					fmt.Sprintf("resources for %v are not specified", t),
				)
				return
			}

			tflog.Debug(ctx, fmt.Sprintf("Creating subcluster for %v", t))
			op, err := sdk.WrapOperation(
				sdk.MDB().SPQR().Cluster().AddSubcluster(ctx, &spqr.AddSubclusterRequest{
					ClusterId: cid,
					HostSpecs: specs,
					Resources: res,
				}),
			)
			if err != nil {
				diag.AddError(
					"Failed to create hosts",
					fmt.Sprintf("Error while requesting API to create host ShardedPostgresql cluster %q: %s", cid, err.Error()),
				)
				return
			}

			if err = op.Wait(ctx); err != nil {
				diag.AddError(
					"Failed to create hosts",
					fmt.Sprintf("Error while waiting for operation %q to create host ShardedPostgresql cluster %q: %s", op.Id(), cid, err.Error()),
				)
				return
			}
		}
	}

	// Add new hosts to existing subclusters
	for _, spec := range addClusterHosts {
		op, err := sdk.WrapOperation(
			sdk.MDB().SPQR().Cluster().AddHosts(ctx, &spqr.AddClusterHostsRequest{
				ClusterId: cid,
				HostSpecs: []*spqr.HostSpec{spec},
			}),
		)
		if err != nil {
			diag.AddError(
				"Failed to create hosts",
				fmt.Sprintf("Error while requesting API to create host ShardedPostgresql cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to create hosts",
				fmt.Sprintf("Error while waiting for operation %q to create host ShardedPostgresql cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

func (p *ShardedPostgreSQLAPI) CreateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*spqr.HostSpec) {
	panic("never should be called")
}

func (p *ShardedPostgreSQLAPI) UpdateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*spqr.UpdateHostSpec) {
	for _, spec := range specs {
		request := &spqr.UpdateClusterHostsRequest{
			ClusterId: cid,
			UpdateHostSpecs: []*spqr.UpdateHostSpec{
				spec,
			},
		}
		op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending ShardedPostgresql cluster update hosts request: %+v", request)
			return sdk.MDB().SPQR().Cluster().UpdateHosts(ctx, request)
		})
		if err != nil {
			diag.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while requesting API to update host ShardedPostgresql cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while waiting for operation %q to update host ShardedPostgresql cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

func (p *ShardedPostgreSQLAPI) DeleteHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, fqdns []string) {
	for _, fqdn := range fqdns {
		op, err := sdk.WrapOperation(
			sdk.MDB().SPQR().Cluster().DeleteHosts(ctx, &spqr.DeleteClusterHostsRequest{
				ClusterId: cid,
				HostNames: []string{fqdn},
			}),
		)
		if err != nil {
			diag.AddError(
				"Failed to delete hosts",
				fmt.Sprintf("Error while requesting API to delete host ShardedPostgresql cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to delete hosts",
				fmt.Sprintf("Error while waiting for operation %q to delete host ShardedPostgresql cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

func (p *ShardedPostgreSQLAPI) ListHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*spqr.Host {
	attempts := 7
	return p.retryListSPQRHostsInner(ctx, sdk, diags, cid, 0, attempts, func(hosts []*spqr.Host) bool {
		return true
	})
}

func (r *ShardedPostgreSQLAPI) retryListSPQRHostsInner(
	ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, attempt int, maxAttempt int, condition func([]*spqr.Host) bool) []*spqr.Host {
	log.Printf("[DEBUG] Try ListSPQRHosts, attempt: %d", attempt)
	hosts := r.listHostsOnce(ctx, sdk, diags, cid)
	if diags.HasError() {
		return nil
	}
	if condition(hosts) || maxAttempt <= attempt {
		return hosts // We tried to do our best
	}

	timeout := int(math.Pow(2, float64(attempt)))
	log.Printf("[DEBUG] Condition failed, waiting %ds before the next attempt", timeout)
	time.Sleep(time.Second * time.Duration(timeout))

	return r.retryListSPQRHostsInner(ctx, sdk, diags, cid, attempt+1, maxAttempt, condition)
}

// Do not use. Use ListHosts instead
func (p *ShardedPostgreSQLAPI) listHostsOnce(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*spqr.Host {
	hosts := []*spqr.Host{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().SPQR().Cluster().ListHosts(ctx, &spqr.ListClusterHostsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to List ShardedPostgresql Hosts",
				"Error while requesting API to get ShardedPostgresql host:"+err.Error(),
			)
			return nil
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts
}
