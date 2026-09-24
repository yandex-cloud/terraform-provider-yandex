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

type MysqlAPI struct{}

// ==============================================================================
//                                     HOST
// ==============================================================================

// retry with 1, 2, 4, 8, 16, 32, 64, 128 seconds if no succeess
// while at least one host is unknown and there is no master
func (r *MysqlAPI) ListHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*mysql.Host {
	attempts := 7

	return r.retryListMySQLHostsInner(ctx, sdk, diags, cid, 0, attempts, func(hosts []*mysql.Host) bool {
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

func (r *MysqlAPI) retryListMySQLHostsInner(
	ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, attempt int, maxAttempt int, condition func([]*mysql.Host) bool) []*mysql.Host {
	log.Printf("[DEBUG] Try ListMySQLHosts, attempt: %d", attempt)
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

	return r.retryListMySQLHostsInner(ctx, sdk, diags, cid, attempt+1, maxAttempt, condition)
}

// Do not use. Use ListHosts instead
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
