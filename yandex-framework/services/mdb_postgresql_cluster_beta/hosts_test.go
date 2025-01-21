package mdb_postgresql_cluster_beta

import (
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func Test_hostsDiffToCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		planHosts map[string]Host
		apiHosts  map[string]Host
		expected  map[string]*postgresql.HostSpec
	}{
		{
			name: "create new host",
			planHosts: map[string]Host{
				"host1": {
					Zone:              types.StringValue("zone1"),
					SubnetId:          types.StringValue("subnet1"),
					AssignPublicIp:    types.BoolValue(true),
					ReplicationSource: types.StringValue("source1"),
				},
			},
			apiHosts: map[string]Host{},
			expected: map[string]*postgresql.HostSpec{
				"host1": {
					ZoneId:            "zone1",
					SubnetId:          "subnet1",
					AssignPublicIp:    true,
					ReplicationSource: "source1",
				},
			},
		},
		{
			name: "no new hosts to create",
			planHosts: map[string]Host{
				"host1": {
					FQDN:              types.StringValue("fqdn1"),
					Zone:              types.StringValue("zone1"),
					SubnetId:          types.StringValue("subnet1"),
					AssignPublicIp:    types.BoolValue(true),
					ReplicationSource: types.StringValue("source1"),
				},
			},
			apiHosts: map[string]Host{
				"host1": {
					FQDN:              types.StringValue("fqdn1"),
					Zone:              types.StringValue("zone1"),
					SubnetId:          types.StringValue("subnet1"),
					AssignPublicIp:    types.BoolValue(true),
					ReplicationSource: types.StringValue("source1"),
				},
			},
			expected: map[string]*postgresql.HostSpec{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toCreate, _, _ := hostsDiff(tt.planHosts, tt.apiHosts)
			if !reflect.DeepEqual(toCreate, tt.expected) {
				t.Errorf("hostsDiff() gotToCreate = %v, want %v", toCreate, tt.expected)
			}
		})
	}
}

func Test_hostsDiffToUpdate(t *testing.T) {
	tests := []struct {
		name      string
		planHosts map[string]Host
		apiHosts  map[string]Host
		expected  []*postgresql.UpdateHostSpec
	}{
		{
			name: "update existing host",
			planHosts: map[string]Host{
				"host1": {
					FQDN:              types.StringValue("fqdn1"),
					AssignPublicIp:    types.BoolValue(true),
					ReplicationSource: types.StringValue("source1"),
				},
			},
			apiHosts: map[string]Host{
				"host1": {
					FQDN:              types.StringValue("fqdn1"),
					AssignPublicIp:    types.BoolValue(false),
					ReplicationSource: types.StringValue("source2"),
				},
			},
			expected: []*postgresql.UpdateHostSpec{
				{
					HostName:          "fqdn1",
					UpdateMask:        &fieldmaskpb.FieldMask{Paths: []string{"assign_public_ip", "replication_source"}},
					AssignPublicIp:    true,
					ReplicationSource: "source1",
				},
			},
		},
		{
			name: "no hosts to update",
			planHosts: map[string]Host{
				"host1": {
					FQDN:              types.StringValue("fqdn1"),
					AssignPublicIp:    types.BoolValue(true),
					ReplicationSource: types.StringValue("source1"),
				},
			},
			apiHosts: map[string]Host{
				"host1": {
					FQDN:              types.StringValue("fqdn1"),
					AssignPublicIp:    types.BoolValue(true),
					ReplicationSource: types.StringValue("source1"),
				},
			},
			expected: []*postgresql.UpdateHostSpec{},
		},
		{
			name: "update multiple hosts",
			planHosts: map[string]Host{
				"na": {
					Zone:              types.StringValue("ru-central1-a"),
					SubnetId:          types.StringValue("bucvu6sbipulaibm5pv9"),
					AssignPublicIp:    types.BoolValue(false),
					FQDN:              types.StringValue("rc1a-ezr3kr2cjc06gig1.mdb.cloud-preprod.yandex.net"),
					ReplicationSource: types.StringValue(""),
				},
				"nb": {
					Zone:              types.StringValue("ru-central1-b"),
					SubnetId:          types.StringValue("blt1km2hioumn14dsuoc"),
					AssignPublicIp:    types.BoolValue(true),
					FQDN:              types.StringValue("rc1b-jsr1v6dz7mhw7ltz.mdb.cloud-preprod.yandex.net"),
					ReplicationSource: types.StringValue(""),
				},
				"nd": {
					Zone:              types.StringValue("ru-central1-d"),
					SubnetId:          types.StringValue("fqsfn494qmtoc1rjoldh"),
					AssignPublicIp:    types.BoolValue(false),
					FQDN:              types.StringValue("rc1d-altct5nyf74ivs1a.mdb.cloud-preprod.yandex.net"),
					ReplicationSource: types.StringValue(""),
				},
			},
			apiHosts: map[string]Host{
				"na": {
					Zone:              types.StringValue("ru-central1-a"),
					SubnetId:          types.StringValue("bucvu6sbipulaibm5pv9"),
					AssignPublicIp:    types.BoolValue(true),
					FQDN:              types.StringValue("rc1a-ezr3kr2cjc06gig1.mdb.cloud-preprod.yandex.net"),
					ReplicationSource: types.StringValue(""),
				},
				"nb": {
					Zone:              types.StringValue("ru-central1-b"),
					SubnetId:          types.StringValue("blt1km2hioumn14dsuoc"),
					AssignPublicIp:    types.BoolValue(false),
					FQDN:              types.StringValue("rc1b-jsr1v6dz7mhw7ltz.mdb.cloud-preprod.yandex.net"),
					ReplicationSource: types.StringValue(""),
				},
				"nd": {
					Zone:              types.StringValue("ru-central1-d"),
					SubnetId:          types.StringValue("fqsfn494qmtoc1rjoldh"),
					AssignPublicIp:    types.BoolValue(true),
					FQDN:              types.StringValue("rc1d-altct5nyf74ivs1a.mdb.cloud-preprod.yandex.net"),
					ReplicationSource: types.StringValue(""),
				},
			},
			expected: []*postgresql.UpdateHostSpec{
				{
					HostName:          "rc1a-ezr3kr2cjc06gig1.mdb.cloud-preprod.yandex.net",
					UpdateMask:        &fieldmaskpb.FieldMask{Paths: []string{"assign_public_ip", "replication_source"}},
					AssignPublicIp:    false,
					ReplicationSource: "",
				},
				{
					HostName:          "rc1b-jsr1v6dz7mhw7ltz.mdb.cloud-preprod.yandex.net",
					UpdateMask:        &fieldmaskpb.FieldMask{Paths: []string{"assign_public_ip", "replication_source"}},
					AssignPublicIp:    true,
					ReplicationSource: "",
				},
				{
					HostName:          "rc1d-altct5nyf74ivs1a.mdb.cloud-preprod.yandex.net",
					UpdateMask:        &fieldmaskpb.FieldMask{Paths: []string{"assign_public_ip", "replication_source"}},
					AssignPublicIp:    false,
					ReplicationSource: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, toUpdate, _ := hostsDiff(tt.planHosts, tt.apiHosts)

			sort.Slice(toUpdate, func(i, j int) bool {
				return toUpdate[i].HostName < toUpdate[j].HostName
			})

			if !reflect.DeepEqual(toUpdate, tt.expected) {
				t.Errorf("hostsDiff() gotToUpdate = %v, want %v", toUpdate, tt.expected)
			}
		})
	}
}

func Test_hostsDiffToDelete(t *testing.T) {
	tests := []struct {
		name      string
		planHosts map[string]Host
		apiHosts  map[string]Host
		expected  []string
	}{
		{
			name:      "delete existing host",
			planHosts: map[string]Host{},
			apiHosts: map[string]Host{
				"host1": {
					FQDN: types.StringValue("fqdn1"),
				},
			},
			expected: []string{"fqdn1"},
		},
		{
			name: "no hosts to delete",
			planHosts: map[string]Host{
				"host1": {
					FQDN: types.StringValue("fqdn1"),
				},
			},
			apiHosts: map[string]Host{
				"host1": {
					FQDN: types.StringValue("fqdn1"),
				},
			},
			expected: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, toDelete := hostsDiff(tt.planHosts, tt.apiHosts)
			if !reflect.DeepEqual(toDelete, tt.expected) {
				t.Errorf("hostsDiff() gotToDelete = %v, want %v", toDelete, tt.expected)
			}
		})
	}
}
