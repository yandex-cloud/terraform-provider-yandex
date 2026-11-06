package mdb_mysql_cluster_v2

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	mysqlsdk "github.com/yandex-cloud/go-sdk/services/mdb/mysql/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

const defaultMDBPageSize = 1000

var mysqlApi = MysqlAPI{}

type MysqlAPI struct {
	listHostsMaxAttempt *int // nil: defaultListHostsMaxAttempt
}

// ==============================================================================
//                                     HOST
// ==============================================================================

const defaultListHostsMaxAttempt = 7

func (r *MysqlAPI) ListHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*mysql.Host {
	return retryListHosts(ctx, diags, func(ctx context.Context) []*mysql.Host {
		return r.listHostsOnce(ctx, sdk, diags, cid)
	}, r.maxAttempt(), retryBackoff)
}

func (r *MysqlAPI) maxAttempt() int {
	if r.listHostsMaxAttempt == nil {
		return defaultListHostsMaxAttempt
	}
	return *r.listHostsMaxAttempt
}

// A STOPPED cluster reports ROLE_UNKNOWN for every host, so waiting for MASTER there would only spend the whole retry budget.
func hostsAPIForRead(cluster *mysql.Cluster) *MysqlAPI {
	api := &MysqlAPI{}
	if cluster.GetStatus() == mysql.Cluster_STOPPED {
		noWait := 0
		api.listHostsMaxAttempt = &noWait
	}
	return api
}

func retryBackoff(attempt int) time.Duration {
	return time.Second * time.Duration(int(math.Pow(2, float64(attempt))))
}

func retryListHosts(ctx context.Context, diags *diag.Diagnostics, list func(context.Context) []*mysql.Host, maxAttempt int, backoff func(int) time.Duration) []*mysql.Host {
	for attempt := 0; ; attempt++ {
		log.Printf("[DEBUG] Try ListMySQLHosts, attempt: %d", attempt)
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
				"Failed to List MySQL Hosts",
				"Context is done while waiting for MySQL host roles: "+ctx.Err().Error(),
			)
			return nil
		case <-time.After(wait):
		}
	}
}

// Unknown roles are not waited for: the role is not stored in the state, and a host that stays ROLE_UNKNOWN would not get one by waiting.
func hasMaster(hosts []*mysql.Host) bool {
	for _, host := range hosts {
		if host.Role == mysql.Host_MASTER {
			return true
		}
	}
	return false
}

func (r *MysqlAPI) listHostsOnce(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*mysql.Host {
	hosts := []*mysql.Host{}
	pageToken := ""

	for {
		resp, err := mysqlsdk.NewClusterClient(sdk).ListHosts(ctx, &mysql.ListClusterHostsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to List MySQL Hosts",
				"Error while requesting API to get MySQL host:"+err.Error(),
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

func (r *MysqlAPI) CreateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*mysql.HostSpec, opts struct{}) {
	for _, spec := range specs {
		op, err :=
			mysqlsdk.NewClusterClient(sdk).AddHosts(ctx, &mysql.AddClusterHostsRequest{
				ClusterId: cid,
				HostSpecs: []*mysql.HostSpec{spec},
			})
		if err != nil {
			diag.AddError(
				"Failed to create hosts",
				fmt.Sprintf("Error while requesting API to create host MySQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if _, err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to create hosts",
				fmt.Sprintf("Error while waiting for operation %q to create host MySQL cluster %q: %s", op.ID(), cid, err.Error()),
			)
			return
		}
	}
}

func (r *MysqlAPI) UpdateHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, specs []*mysql.UpdateHostSpec) {
	for _, spec := range specs {
		request := &mysql.UpdateClusterHostsRequest{
			ClusterId: cid,
			UpdateHostSpecs: []*mysql.UpdateHostSpec{
				spec,
			},
		}
		op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*mysqlsdk.ClusterUpdateHostsOperation, error) {
			log.Printf("[DEBUG] Sending MySQL cluster update hosts request: %+v", request)
			return mysqlsdk.NewClusterClient(sdk).UpdateHosts(ctx, request)
		})
		if err != nil {
			diag.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while requesting API to update host MySQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if _, err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while waiting for operation %q to update host MySQL cluster %q: %s", op.ID(), cid, err.Error()),
			)
			return
		}
	}
}

func (r *MysqlAPI) DeleteHosts(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, fqdns []string) {
	for _, fqdn := range fqdns {
		op, err :=
			mysqlsdk.NewClusterClient(sdk).DeleteHosts(ctx, &mysql.DeleteClusterHostsRequest{
				ClusterId: cid,
				HostNames: []string{fqdn},
			})
		if err != nil {
			diag.AddError(
				"Failed to delete hosts",
				fmt.Sprintf("Error while requesting API to delete host MySQL cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if _, err = op.Wait(ctx); err != nil {
			diag.AddError(
				"Failed to delete hosts",
				fmt.Sprintf("Error while waiting for operation %q to delete host MySQL cluster %q: %s", op.ID(), cid, err.Error()),
			)
			return
		}
	}
}

// ==============================================================================
//                                 CLUSTER
// ==============================================================================

func (r *MysqlAPI) GetCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) *mysql.Cluster {
	db, err := mysqlsdk.NewClusterClient(sdk).Get(ctx, &mysql.GetClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diags.AddError(
			"Failed to read resource",
			fmt.Sprintf("Error while requesting API to read MySQL cluster %q: %s", cid, err.Error()),
		)
		return nil
	}
	return db
}

func (r *MysqlAPI) DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) {
	op, err := mysqlsdk.NewClusterClient(sdk).Delete(ctx, &mysql.DeleteClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete MySQL cluster %q: %s", cid, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Deleting MySQL Cluster", map[string]any{"cluster_id": cid})

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete MySQL cluster %q: %s", op.ID(), cid, err.Error()),
		)
	}
}

func (r *MysqlAPI) CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *mysql.CreateClusterRequest) string {
	op, err := mysqlsdk.NewClusterClient(sdk).Create(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create MySQL cluster: %s", err.Error()),
		)
		return ""
	}

	md := op.Metadata()

	tflog.Debug(ctx, "Creating MySQL Cluster", map[string]any{"request_body": req})

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create MySQL cluster: %s", op.ID(), err.Error()),
		)
		return ""
	}

	return md.ClusterId
}

func (p *MysqlAPI) RestoreCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *mysql.RestoreClusterRequest) string {
	op, err := mysqlsdk.NewClusterClient(sdk).Restore(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to restore resource",
			fmt.Sprintf("Error while requesting API to restore MySQL cluster: %s", err.Error()),
		)
		return ""
	}

	md := op.Metadata()

	tflog.Debug(ctx, "Restoring MySQL cluster", map[string]any{"request_body": req})

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to restore resource",
			fmt.Sprintf("Error while waiting for operation %q to restore MySQL cluster: %s", op.ID(), err.Error()),
		)
		return ""
	}

	return md.ClusterId
}

func (r *MysqlAPI) UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *mysql.UpdateClusterRequest) {

	if req == nil || len(req.UpdateMask.Paths) == 0 {
		return
	}

	op, err := mysqlsdk.NewClusterClient(sdk).Update(ctx, req)
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update MySQL cluster: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Updating MySQL Cluster", map[string]any{"request_body": req})

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update MySQL cluster: %s", op.ID(), err.Error()),
		)
		return
	}
}
