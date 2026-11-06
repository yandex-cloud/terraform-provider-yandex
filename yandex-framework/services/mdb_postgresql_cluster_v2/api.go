package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	postgresqlsdk "github.com/yandex-cloud/go-sdk/services/mdb/postgresql/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

const defaultMDBPageSize = 1000

var postgresqlApi = PostgresqlAPI{}

type PostgresqlAPI struct {
	listHostsMaxAttempt *int // nil: defaultListHostsMaxAttempt
}

// ==============================================================================
//                                     HOST
// ==============================================================================

const defaultListHostsMaxAttempt = 7

func (p *PostgresqlAPI) ListHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*postgresql.Host {
	return retryListHosts(ctx, diags, func(ctx context.Context) []*postgresql.Host {
		return p.listHostsOnce(ctx, sdk, diags, cid)
	}, p.maxAttempt(), retryBackoff)
}

func (p *PostgresqlAPI) maxAttempt() int {
	if p.listHostsMaxAttempt == nil {
		return defaultListHostsMaxAttempt
	}
	return *p.listHostsMaxAttempt
}

func hostsAPIForRead(cluster *postgresql.Cluster) *PostgresqlAPI {
	api := &PostgresqlAPI{}
	if cluster.GetStatus() == postgresql.Cluster_STOPPED {
		noWait := 0
		api.listHostsMaxAttempt = &noWait
	}
	return api
}

func retryBackoff(attempt int) time.Duration {
	return time.Second * time.Duration(int(math.Pow(2, float64(attempt))))
}

func retryListHosts(ctx context.Context, diags *diag.Diagnostics, list func(context.Context) []*postgresql.Host, maxAttempt int, backoff func(int) time.Duration) []*postgresql.Host {
	for attempt := 0; ; attempt++ {
		log.Printf("[DEBUG] Try ListPostgreSQLHosts, attempt: %d", attempt)
		hosts := list(ctx)
		if diags.HasError() {
			return nil
		}
		if hasMaster(hosts) || maxAttempt <= attempt {
			return hosts
		}

		wait := backoff(attempt)
		log.Printf("[DEBUG] Condition failed, waiting %s before the next attempt", wait)
		select {
		case <-ctx.Done():
			diags.AddError(
				"Failed to List PostgreSQL Hosts",
				"Context is done while waiting for PostgreSQL host roles: "+ctx.Err().Error(),
			)
			return nil
		case <-time.After(wait):
		}
	}
}

func hasMaster(hosts []*postgresql.Host) bool {
	for _, host := range hosts {
		if host.Role == postgresql.Host_MASTER {
			return true
		}
	}
	return false
}

func (p *PostgresqlAPI) listHostsOnce(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*postgresql.Host {
	hosts := []*postgresql.Host{}
	pageToken := ""

	for {
		resp, err := postgresqlsdk.NewClusterClient(sdk).ListHosts(ctx, &postgresql.ListClusterHostsRequest{
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

func (p *PostgresqlAPI) CreateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*postgresql.HostSpec, opts struct{}) {
	for _, spec := range specs {
		op, err :=
			postgresqlsdk.NewClusterClient(sdk).AddHosts(ctx, &postgresql.AddClusterHostsRequest{
				ClusterId: cid,
				HostSpecs: []*postgresql.HostSpec{spec},
			})
		if err != nil {
			diag.AddError(
				"Failed to create hosts",
				fmt.Sprintf("Error while requesting API to create host PostgreSQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if _, err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to create hosts",
				fmt.Sprintf("Error while waiting for operation %q to create host PostgreSQL cluster %q: %s", op.ID(), cid, err.Error()),
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
		op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*postgresqlsdk.ClusterUpdateHostsOperation, error) {
			log.Printf("[DEBUG] Sending PostgreSQL cluster update hosts request: %+v", request)
			return postgresqlsdk.NewClusterClient(sdk).UpdateHosts(ctx, request)
		})
		if err != nil {
			diag.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while requesting API to update host PostgreSQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if _, err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while waiting for operation %q to update host PostgreSQL cluster %q: %s", op.ID(), cid, err.Error()),
			)
			return
		}
	}
}

func (p *PostgresqlAPI) DeleteHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, fqdns []string) {
	for _, fqdn := range fqdns {
		op, err :=
			postgresqlsdk.NewClusterClient(sdk).DeleteHosts(ctx, &postgresql.DeleteClusterHostsRequest{
				ClusterId: cid,
				HostNames: []string{fqdn},
			})
		if err != nil {
			diag.AddError(
				"Failed to delete hosts",
				fmt.Sprintf("Error while requesting API to delete host PostgreSQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if _, err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to delete hosts",
				fmt.Sprintf("Error while waiting for operation %q to delete host PostgreSQL cluster %q: %s", op.ID(), cid, err.Error()),
			)
			return
		}
	}
}

// ==============================================================================
//                                 CLUSTER
// ==============================================================================

func (p *PostgresqlAPI) GetCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) *postgresql.Cluster {
	db, err := postgresqlsdk.NewClusterClient(sdk).Get(ctx, &postgresql.GetClusterRequest{
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
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*postgresqlsdk.ClusterDeleteOperation, error) {
		return postgresqlsdk.NewClusterClient(sdk).Delete(ctx, &postgresql.DeleteClusterRequest{
			ClusterId: cid,
		})
	})

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete PostgreSQL cluster %q: %s", cid, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Deleting PostgreSQL Cluster", map[string]any{"cluster_id": cid})

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete PostgreSQL cluster %q: %s", op.ID(), cid, err.Error()),
		)
	}
}

func (p *PostgresqlAPI) CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *postgresql.CreateClusterRequest) string {
	op, err := postgresqlsdk.NewClusterClient(sdk).Create(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create PostgreSQL cluster: %s", err.Error()),
		)
		return ""
	}

	md := op.Metadata()

	tflog.Debug(ctx, "Creating PostgreSQL Cluster", map[string]any{"request_body": req})

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create PostgreSQL cluster: %s", op.ID(), err.Error()),
		)
		return md.ClusterId
	}

	return md.ClusterId
}

func (p *PostgresqlAPI) RestoreCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *postgresql.RestoreClusterRequest) string {
	op, err := postgresqlsdk.NewClusterClient(sdk).Restore(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to restore resource from backup",
			fmt.Sprintf("Error while requesting API to restore PostgreSQL cluster from backup: %s", err.Error()),
		)
		return ""
	}

	md := op.Metadata()

	tflog.Debug(ctx, "Restoring PostgreSQL Cluster from backup", map[string]any{"request_body": req})

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to restore resource from backup",
			fmt.Sprintf("Error while waiting for operation %q to restore PostgreSQL cluster from backup: %s", op.ID(), err.Error()),
		)
		return md.ClusterId
	}

	return md.ClusterId
}

func (p *PostgresqlAPI) UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *postgresql.UpdateClusterRequest) {
	if req == nil || len(req.UpdateMask.Paths) == 0 {
		return
	}

	op, err := postgresqlsdk.NewClusterClient(sdk).Update(ctx, req)
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update PostgreSQL cluster: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Updating PostgreSQL Cluster", map[string]any{"request_body": req})
	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update PostgreSQL cluster: %s", op.ID(), err.Error()),
		)
		return
	}
}
