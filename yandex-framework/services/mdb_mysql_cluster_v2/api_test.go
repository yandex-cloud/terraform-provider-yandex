package mdb_mysql_cluster_v2

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
)

func TestHasMaster(t *testing.T) {
	cases := []struct {
		name  string
		roles []mysql.Host_Role
		want  bool
	}{
		{"no hosts", nil, false},
		{"unknown only", []mysql.Host_Role{mysql.Host_ROLE_UNKNOWN}, false},
		{"replica and unknown", []mysql.Host_Role{mysql.Host_REPLICA, mysql.Host_ROLE_UNKNOWN}, false},
		{"master and unknown", []mysql.Host_Role{mysql.Host_MASTER, mysql.Host_ROLE_UNKNOWN}, true},
		{"master and replica", []mysql.Host_Role{mysql.Host_REPLICA, mysql.Host_MASTER}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hosts := make([]*mysql.Host, 0, len(tc.roles))
			for _, role := range tc.roles {
				hosts = append(hosts, &mysql.Host{Role: role})
			}
			assert.Equal(t, tc.want, hasMaster(hosts))
		})
	}
}

func TestRetryBackoff(t *testing.T) {
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second, 64 * time.Second}
	for attempt, d := range want {
		assert.Equal(t, d, retryBackoff(attempt), "attempt %d", attempt)
	}
}

func TestHostsAPIForRead(t *testing.T) {
	cases := []struct {
		name   string
		status mysql.Cluster_Status
		want   int
	}{
		{"stopped cluster is read without waiting", mysql.Cluster_STOPPED, 0},
		{"unknown status waits for MASTER", mysql.Cluster_STATUS_UNKNOWN, defaultListHostsMaxAttempt},
		{"running cluster waits for MASTER", mysql.Cluster_RUNNING, defaultListHostsMaxAttempt},
		{"updating cluster waits for MASTER", mysql.Cluster_UPDATING, defaultListHostsMaxAttempt},
		{"stopping cluster waits for MASTER", mysql.Cluster_STOPPING, defaultListHostsMaxAttempt},
		{"starting cluster waits for MASTER", mysql.Cluster_STARTING, defaultListHostsMaxAttempt},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, hostsAPIForRead(&mysql.Cluster{Status: tc.status}).maxAttempt())
		})
	}
	assert.Equal(t, defaultListHostsMaxAttempt, hostsAPIForRead(nil).maxAttempt())
	assert.Equal(t, defaultListHostsMaxAttempt, mysqlApi.maxAttempt(), "the shared API value must keep the default budget")
}

func TestRetryListHosts(t *testing.T) {
	unknownHosts := []*mysql.Host{{Name: "a", Role: mysql.Host_ROLE_UNKNOWN}}
	masterHosts := []*mysql.Host{{Name: "a", Role: mysql.Host_MASTER}}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name       string
		ctx        context.Context
		hosts      []*mysql.Host
		listErr    bool
		maxAttempt int
		wait       time.Duration
		wantWaits  []int
		wantHosts  []*mysql.Host
		wantErr    bool
		wantCalls  int
	}{
		{name: "single attempt returns hosts without roles", hosts: unknownHosts, wantHosts: unknownHosts, wantCalls: 1},
		{name: "single attempt returns nil on list error", listErr: true, maxAttempt: defaultListHostsMaxAttempt, wantErr: true, wantCalls: 1},
		{name: "stops as soon as a MASTER is listed", hosts: masterHosts, maxAttempt: defaultListHostsMaxAttempt, wantHosts: masterHosts, wantCalls: 1},
		{name: "spends the whole budget and returns the last result", hosts: unknownHosts, maxAttempt: 2, wantWaits: []int{0, 1}, wantHosts: unknownHosts, wantCalls: 3},
		{name: "wait ends when context is done", ctx: canceled, hosts: unknownHosts, maxAttempt: defaultListHostsMaxAttempt, wait: time.Hour, wantWaits: []int{0}, wantErr: true, wantCalls: 1},
		{name: "zero backoff does not stop on a live context", hosts: unknownHosts, maxAttempt: 1, wantWaits: []int{0}, wantHosts: unknownHosts, wantCalls: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			var diags diag.Diagnostics
			calls := 0
			list := func(context.Context) []*mysql.Host {
				calls++
				if tc.listErr {
					diags.AddError("list failed", "")
				}
				return tc.hosts
			}
			var waits []int
			backoff := func(attempt int) time.Duration {
				waits = append(waits, attempt)
				return tc.wait
			}

			hosts := retryListHosts(ctx, &diags, list, tc.maxAttempt, backoff)

			assert.Equal(t, tc.wantErr, diags.HasError())
			assert.Equal(t, tc.wantHosts, hosts)
			assert.Equal(t, tc.wantCalls, calls)
			assert.Equal(t, tc.wantWaits, waits)
		})
	}
}
