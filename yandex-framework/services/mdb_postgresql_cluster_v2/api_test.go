package mdb_postgresql_cluster_v2

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func TestHasMaster(t *testing.T) {
	cases := []struct {
		name  string
		roles []postgresql.Host_Role
		want  bool
	}{
		{"no hosts", nil, false},
		{"unknown only", []postgresql.Host_Role{postgresql.Host_ROLE_UNKNOWN}, false},
		{"replica and unknown", []postgresql.Host_Role{postgresql.Host_REPLICA, postgresql.Host_ROLE_UNKNOWN}, false},
		{"master and unknown", []postgresql.Host_Role{postgresql.Host_MASTER, postgresql.Host_ROLE_UNKNOWN}, true},
		{"master and replica", []postgresql.Host_Role{postgresql.Host_REPLICA, postgresql.Host_MASTER}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hosts := make([]*postgresql.Host, 0, len(tc.roles))
			for _, role := range tc.roles {
				hosts = append(hosts, &postgresql.Host{Role: role})
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
		status postgresql.Cluster_Status
		want   int
	}{
		{"stopped cluster is read without waiting", postgresql.Cluster_STOPPED, 0},
		{"unknown status waits for MASTER", postgresql.Cluster_STATUS_UNKNOWN, defaultListHostsMaxAttempt},
		{"running cluster waits for MASTER", postgresql.Cluster_RUNNING, defaultListHostsMaxAttempt},
		{"updating cluster waits for MASTER", postgresql.Cluster_UPDATING, defaultListHostsMaxAttempt},
		{"stopping cluster waits for MASTER", postgresql.Cluster_STOPPING, defaultListHostsMaxAttempt},
		{"starting cluster waits for MASTER", postgresql.Cluster_STARTING, defaultListHostsMaxAttempt},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, hostsAPIForRead(&postgresql.Cluster{Status: tc.status}).maxAttempt())
		})
	}
	assert.Equal(t, defaultListHostsMaxAttempt, hostsAPIForRead(nil).maxAttempt())
	assert.Equal(t, defaultListHostsMaxAttempt, postgresqlApi.maxAttempt(), "the shared API value must keep the default budget")
}

func TestRetryListHosts(t *testing.T) {
	unknownHosts := []*postgresql.Host{{Name: "a", Role: postgresql.Host_ROLE_UNKNOWN}}
	masterHosts := []*postgresql.Host{{Name: "a", Role: postgresql.Host_MASTER}}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name       string
		ctx        context.Context
		hosts      []*postgresql.Host
		listErr    bool
		maxAttempt int
		wait       time.Duration
		wantWaits  []int
		wantHosts  []*postgresql.Host
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
			list := func(context.Context) []*postgresql.Host {
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
