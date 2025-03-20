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

var postgresqlApi = PostgresqlAPI{}

type PostgresqlAPI struct{}

// ==============================================================================
//                                     HOST
// ==============================================================================

// retry with 1, 2, 4, 8, 16, 32, 64, 128 seconds if no succeess
// while at least one host is unknown and there is no master
func (p *PostgresqlAPI) ListHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*postgresql.Host {
	attempts := 7
	return p.retryListPostgreSQLHostsInner(ctx, sdk, diags, cid, 0, attempts, func(hosts []*postgresql.Host) bool {
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

func (r *PostgresqlAPI) retryListPostgreSQLHostsInner(
	ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, attempt int, maxAttempt int, condition func([]*postgresql.Host) bool) []*postgresql.Host {
	log.Printf("[DEBUG] Try ListPostgreSQLHosts, attempt: %d", attempt)
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

	return r.retryListPostgreSQLHostsInner(ctx, sdk, diags, cid, attempt+1, maxAttempt, condition)
}

// Do not use. Use ListHosts instead
func (p *PostgresqlAPI) listHostsOnce(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*postgresql.Host {
	hosts := []*postgresql.Host{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().PostgreSQL().Cluster().ListHosts(ctx, &postgresql.ListClusterHostsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to List PostgreSQL Hosts",
				"Error while requesting API to get PostgreSQL host:"+err.Error(),
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

func (p *PostgresqlAPI) CreateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*postgresql.HostSpec) {
	for _, spec := range specs {
		op, err := sdk.WrapOperation(
			sdk.MDB().PostgreSQL().Cluster().AddHosts(ctx, &postgresql.AddClusterHostsRequest{
				ClusterId: cid,
				HostSpecs: []*postgresql.HostSpec{spec},
			}),
		)
		if err != nil {
			diag.AddError(
				"Failed to create hosts",
				fmt.Sprintf("Error while requesting API to create host PostgreSQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to create hosts",
				fmt.Sprintf("Error while waiting for operation %q to create host PostgreSQL cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

func (p *PostgresqlAPI) UpdateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*postgresql.UpdateHostSpec) {
	for _, spec := range specs {
		request := &postgresql.UpdateClusterHostsRequest{
			ClusterId: cid,
			UpdateHostSpecs: []*postgresql.UpdateHostSpec{
				spec,
			},
		}
		op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending PostgreSQL cluster update hosts request: %+v", request)
			return sdk.MDB().PostgreSQL().Cluster().UpdateHosts(ctx, request)
		})
		if err != nil {
			diag.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while requesting API to update host PostgreSQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while waiting for operation %q to update host PostgreSQL cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

func (p *PostgresqlAPI) DeleteHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, fqdns []string) {
	for _, fqdn := range fqdns {
		op, err := sdk.WrapOperation(
			sdk.MDB().PostgreSQL().Cluster().DeleteHosts(ctx, &postgresql.DeleteClusterHostsRequest{
				ClusterId: cid,
				HostNames: []string{fqdn},
			}),
		)
		if err != nil {
			diag.AddError(
				"Failed to delete hosts",
				fmt.Sprintf("Error while requesting API to delete host PostgreSQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to delete hosts",
				fmt.Sprintf("Error while waiting for operation %q to delete host PostgreSQL cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

// ==============================================================================
//                                 CLUSTER
// ==============================================================================

func (p *PostgresqlAPI) GetCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) *postgresql.Cluster {
	db, err := sdk.MDB().PostgreSQL().Cluster().Get(ctx, &postgresql.GetClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diags.AddError(
			"Failed to read resource",
			fmt.Sprintf("Error while requesting API to read PostgreSQL cluster %q: %s", cid, err.Error()),
		)
		return nil
	}
	return db
}

func (p *PostgresqlAPI) DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) {
	op, err := sdk.WrapOperation(sdk.MDB().PostgreSQL().Cluster().Delete(ctx, &postgresql.DeleteClusterRequest{
		ClusterId: cid,
	}))

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete PostgreSQL cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete PostgreSQL cluster %q: %s", op.Id(), cid, err.Error()),
		)
	}
}

func (p *PostgresqlAPI) CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *postgresql.CreateClusterRequest) string {
	op, err := sdk.WrapOperation(sdk.MDB().PostgreSQL().Cluster().Create(ctx, req))
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create PostgreSQL cluster: %s", err.Error()),
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

	md, ok := protoMetadata.(*postgresql.CreateClusterMetadata)
	if !ok {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata", op.Id()),
		)
		return ""
	}

	log.Printf("[DEBUG] Creating PostgreSQL Cluster %q", md.ClusterId)

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create PostgreSQL cluster: %s", op.Id(), err.Error()),
		)
		return ""
	}

	return md.ClusterId
}

func (p *PostgresqlAPI) UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *postgresql.UpdateClusterRequest) {

	if req == nil || len(req.UpdateMask.Paths) == 0 {
		return
	}

	op, err := sdk.WrapOperation(sdk.MDB().PostgreSQL().Cluster().Update(ctx, req))
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update PostgreSQL cluster: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update PostgreSQL cluster: %s", op.Id(), err.Error()),
		)
		return
	}
}
