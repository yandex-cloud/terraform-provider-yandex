package pluginsdk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func TestPGHasMaster(t *testing.T) {
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
			assert.Equal(t, tc.want, pgHasMaster(hosts))
		})
	}
}

func TestPGRetryBackoff(t *testing.T) {
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second, 64 * time.Second}
	for attempt, d := range want {
		assert.Equal(t, d, pgRetryBackoff(attempt), "attempt %d", attempt)
	}
}

func TestPGListHostsMaxAttemptForStatus(t *testing.T) {
	assert.Equal(t, 0, pgListHostsMaxAttemptForStatus(postgresql.Cluster_STOPPED))
	for _, status := range []postgresql.Cluster_Status{
		postgresql.Cluster_STATUS_UNKNOWN,
		postgresql.Cluster_RUNNING,
		postgresql.Cluster_UPDATING,
		postgresql.Cluster_STOPPING,
		postgresql.Cluster_STARTING,
	} {
		assert.Equal(t, pgListHostsMaxAttempt, pgListHostsMaxAttemptForStatus(status), status.String())
	}
}

func TestRetryListPGHosts(t *testing.T) {
	unknownHosts := []*postgresql.Host{{Name: "a", Role: postgresql.Host_ROLE_UNKNOWN}}
	masterHosts := []*postgresql.Host{{Name: "a", Role: postgresql.Host_MASTER}}
	masterAndUnknownHosts := []*postgresql.Host{{Name: "a", Role: postgresql.Host_MASTER}, {Name: "b", Role: postgresql.Host_ROLE_UNKNOWN}}
	listErr := errors.New("list failed")
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name       string
		ctx        context.Context
		hosts      []*postgresql.Host
		listErr    error
		maxAttempt int
		wait       time.Duration
		wantWaits  []int
		wantHosts  []*postgresql.Host
		wantErr    error
		wantCalls  int
	}{
		{name: "single attempt returns hosts without roles", hosts: unknownHosts, wantHosts: unknownHosts, wantCalls: 1},
		{name: "single attempt returns list error", listErr: listErr, wantErr: listErr, wantCalls: 1},
		{name: "stops as soon as a MASTER is listed", hosts: masterHosts, maxAttempt: pgListHostsMaxAttempt, wantHosts: masterHosts, wantCalls: 1},
		{name: "does not wait for the other roles once a MASTER is listed", hosts: masterAndUnknownHosts, maxAttempt: pgListHostsMaxAttempt, wantHosts: masterAndUnknownHosts, wantCalls: 1},
		{name: "spends the whole budget and returns the last result", hosts: unknownHosts, maxAttempt: 2, wantWaits: []int{0, 1}, wantHosts: unknownHosts, wantCalls: 3},
		{name: "list error ends the retries", listErr: listErr, maxAttempt: pgListHostsMaxAttempt, wantErr: listErr, wantCalls: 1},
		{name: "wait ends when context is done", ctx: canceled, hosts: unknownHosts, maxAttempt: pgListHostsMaxAttempt, wait: time.Hour, wantWaits: []int{0}, wantErr: context.Canceled, wantCalls: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			calls := 0
			list := func(context.Context) ([]*postgresql.Host, error) {
				calls++
				return tc.hosts, tc.listErr
			}
			var waits []int
			backoff := func(attempt int) time.Duration {
				waits = append(waits, attempt)
				return tc.wait
			}

			hosts, err := retryListPGHosts(ctx, list, tc.maxAttempt, pgHasMaster, backoff)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.wantHosts, hosts)
			assert.Equal(t, tc.wantCalls, calls)
			assert.Equal(t, tc.wantWaits, waits)
		})
	}
}

func TestPGHostRole(t *testing.T) {
	hosts := []*postgresql.Host{
		{Name: "master.db", Role: postgresql.Host_MASTER},
		{Name: "replica.db", Role: postgresql.Host_REPLICA},
		{Name: "silent.db", Role: postgresql.Host_ROLE_UNKNOWN},
	}
	cases := []struct {
		fqdn string
		want postgresql.Host_Role
	}{
		{"master.db", postgresql.Host_MASTER},
		{"replica.db", postgresql.Host_REPLICA},
		{"silent.db", postgresql.Host_ROLE_UNKNOWN},
		{"missing.db", postgresql.Host_ROLE_UNKNOWN},
	}
	for _, tc := range cases {
		t.Run(tc.fqdn, func(t *testing.T) {
			assert.Equal(t, tc.want, pgHostRole(hosts, tc.fqdn))
		})
	}
}

func TestWaitPGHostRole(t *testing.T) {
	const fqdn = "target.db"
	unknown := []*postgresql.Host{{Name: fqdn, Role: postgresql.Host_ROLE_UNKNOWN}}
	replica := []*postgresql.Host{{Name: fqdn, Role: postgresql.Host_REPLICA}}
	listErr := errors.New("list failed")

	cases := []struct {
		name       string
		role       postgresql.Host_Role
		lists      [][]*postgresql.Host
		listErr    error
		maxAttempt int
		wantRole   postgresql.Host_Role
		wantErr    bool
		wantCalls  int
	}{
		{name: "known role is returned without listing", role: postgresql.Host_MASTER, lists: [][]*postgresql.Host{unknown}, maxAttempt: pgListHostsMaxAttempt, wantRole: postgresql.Host_MASTER},
		{name: "waits until the host reports a role", lists: [][]*postgresql.Host{unknown, replica}, maxAttempt: pgListHostsMaxAttempt, wantRole: postgresql.Host_REPLICA, wantCalls: 2},
		{name: "refuses a host that never reports a role", lists: [][]*postgresql.Host{unknown}, maxAttempt: 1, wantRole: postgresql.Host_ROLE_UNKNOWN, wantErr: true, wantCalls: 2},
		{name: "refuses a host missing from the list", lists: [][]*postgresql.Host{{}}, maxAttempt: 1, wantRole: postgresql.Host_ROLE_UNKNOWN, wantErr: true, wantCalls: 2},
		{name: "list error is returned", lists: [][]*postgresql.Host{unknown}, listErr: listErr, maxAttempt: pgListHostsMaxAttempt, wantRole: postgresql.Host_ROLE_UNKNOWN, wantErr: true, wantCalls: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			list := func(context.Context) ([]*postgresql.Host, error) {
				i := calls
				calls++
				if i >= len(tc.lists) {
					i = len(tc.lists) - 1
				}
				return tc.lists[i], tc.listErr
			}
			noWait := func(int) time.Duration { return 0 }

			role, err := waitPGHostRole(context.Background(), list, fqdn, tc.role, tc.maxAttempt, noWait)

			assert.Equal(t, tc.wantErr, err != nil, "err = %v", err)
			assert.Equal(t, tc.wantRole, role)
			assert.Equal(t, tc.wantCalls, calls)
		})
	}
}
