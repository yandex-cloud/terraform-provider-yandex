package mdbcommon

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

type MockHost struct {
	FQDN             string
	Shard            string
	ParamANotChanged string
	ParamBNotChanged string
	ParamC           int
	ParamD           bool
}

func (m MockHost) GetFQDN() types.String {
	return types.StringValue(m.FQDN)
}
func (m MockHost) GetName() string {
	return m.FQDN
}
func (m MockHost) GetShard() string {
	return m.Shard
}

func (m MockHost) GetShardName() string {
	return m.Shard
}

type MockProtoHostSpec struct {
	ParamANotChanged string
	ParamBNotChanged string
	ParamC           int
	ParamD           bool
}

type MockUpdateSpec struct {
	FQND       string
	ParamC     int
	ParamD     bool
	UpdateMask []string
}

type MockCmpHostService struct {
}

func (m *MockCmpHostService) FullyMatch(plan MockHost, state MockHost) bool {
	return plan.FQDN == state.FQDN && plan.ParamANotChanged == state.ParamANotChanged && plan.ParamBNotChanged == state.ParamBNotChanged &&
		plan.ParamC == state.ParamC && plan.ParamD == state.ParamD
}

func (m *MockCmpHostService) PartialMatch(plan MockHost, state MockHost) bool {
	return plan.FQDN == state.FQDN && plan.ParamANotChanged == state.ParamANotChanged && plan.ParamBNotChanged == state.ParamBNotChanged
}

func (m *MockCmpHostService) GetChanges(plan MockHost, state MockHost) (spec *MockUpdateSpec, diags diag.Diagnostics) {
	if !m.PartialMatch(plan, state) {
		diags.AddError(
			"Wrong state",
			"No change params was changed",
		)
		return nil, diags
	}
	if plan.ParamC == state.ParamC && plan.ParamD == state.ParamD {
		return nil, nil
	}
	return &MockUpdateSpec{
		FQND:       state.FQDN,
		ParamC:     plan.ParamC,
		ParamD:     plan.ParamD,
		UpdateMask: []string{"C", "D"},
	}, diags
}

func (m *MockCmpHostService) ConvertToProto(t MockHost) MockProtoHostSpec {
	return MockProtoHostSpec{
		ParamANotChanged: t.ParamANotChanged,
		ParamBNotChanged: t.ParamBNotChanged,
		ParamC:           t.ParamC,
		ParamD:           t.ParamD,
	}
}

func (m *MockCmpHostService) ConvertFromProto(host MockHost) MockHost {
	return MockHost{
		FQDN:             host.FQDN,
		Shard:            host.Shard,
		ParamANotChanged: host.ParamANotChanged,
		ParamBNotChanged: host.ParamBNotChanged,
		ParamC:           host.ParamC,
		ParamD:           host.ParamD,
	}
}

func TestModifyStateDependsPlan(t *testing.T) {
	hostService := &MockCmpHostService{}

	tests := []struct {
		name           string
		plan           map[string]MockHost
		state          map[string]MockHost
		expectedResult map[string]MockHost
	}{
		{
			name:           "Exact Match Label",
			plan:           map[string]MockHost{"host1": {FQDN: "host1.example.com", ParamC: 1, ParamD: true}},
			state:          map[string]MockHost{"host1": {FQDN: "host1.example.com", ParamC: 1, ParamD: true}},
			expectedResult: map[string]MockHost{"host1": {FQDN: "host1.example.com", ParamC: 1, ParamD: true}},
		},
		{
			name:           "Fully Match with Different Label",
			plan:           map[string]MockHost{"host1": {FQDN: "host1.example.com", ParamC: 1, ParamD: false}},
			state:          map[string]MockHost{"host2": {FQDN: "host1.example.com", ParamC: 1, ParamD: false}},
			expectedResult: map[string]MockHost{"host1": {FQDN: "host1.example.com", ParamC: 1, ParamD: false}},
		},
		{
			name: "Partial Match",
			plan: map[string]MockHost{
				"host1": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B"},
			},
			state: map[string]MockHost{
				"host2": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 3},
			},
			expectedResult: map[string]MockHost{
				"host1": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 3},
			},
		},
		// not stable test because of map order, fix it
		//{
		//	name: "Partial Match But Diff Count Plan",
		//	plan: map[string]MockHost{
		//		"host1": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 1},
		//		"host2": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 2},
		//	},
		//	state: map[string]MockHost{
		//		"host3": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 3},
		//	},
		//	expectedResult: map[string]MockHost{
		//		"host1": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 3},
		//	},
		//},
		// not stable test because of map order, fix it
		//{
		//	name: "Partial Match But Diff Count State",
		//	plan: map[string]MockHost{
		//		"host1": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B"},
		//	},
		//	state: map[string]MockHost{
		//		"host3": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 3},
		//		"host4": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 4},
		//	},
		//	expectedResult: map[string]MockHost{
		//		"host1": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 3},
		//		"host4": {FQDN: "host1.example.com", ParamANotChanged: "A", ParamBNotChanged: "B", ParamC: 4},
		//	},
		//},
		{
			name: "No Match",
			plan: map[string]MockHost{
				"host1": {FQDN: "host1.example.com", ParamC: 5},
			},
			state: map[string]MockHost{
				"host2": {FQDN: "host2.example.com", ParamC: 6},
			},
			expectedResult: map[string]MockHost{
				"host2": {FQDN: "host2.example.com", ParamC: 6},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixedState := modifyStateDependsPlan[MockHost, MockHost, MockProtoHostSpec, MockUpdateSpec](
				hostService, tc.plan, tc.state,
			)
			assert.Equal(t, tc.expectedResult, fixedState)
		})
	}
}

func TestHostsDiff(t *testing.T) {
	hostService := &MockCmpHostService{}

	tests := []struct {
		name             string
		planHosts        map[string]MockHost
		stateHosts       map[string]MockHost
		expectedToCreate []MockProtoHostSpec
		expectedToUpdate []*MockUpdateSpec
		expectedToDelete []string
	}{
		{
			name: "Create and Update",
			planHosts: map[string]MockHost{
				"host1": {FQDN: "host1.example.com", ParamC: 1, ParamD: true},
				"host2": {FQDN: "host2.example.com", ParamC: 3, ParamD: false},
			},
			stateHosts: map[string]MockHost{
				"host2": {FQDN: "host2.example.com", ParamC: 3, ParamD: true},
			},
			expectedToCreate: []MockProtoHostSpec{{ParamC: 1, ParamD: true}},
			expectedToUpdate: []*MockUpdateSpec{{FQND: "host2.example.com", ParamC: 3, ParamD: false, UpdateMask: []string{"C", "D"}}},
			expectedToDelete: nil,
		},
		{
			name: "Delete Hosts",
			planHosts: map[string]MockHost{
				"host1": {FQDN: "host1.example.com", ParamC: 1, ParamD: true},
			},
			stateHosts: map[string]MockHost{
				"host1": {FQDN: "host1.example.com", ParamC: 1, ParamD: true},
				"host3": {FQDN: "host3.example.com", ParamC: 3, ParamD: false},
			},
			expectedToCreate: nil,
			expectedToUpdate: nil,
			expectedToDelete: []string{"host3.example.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			toCreate, toUpdate, toDelete, diags := hostsDiff[MockHost, MockHost, MockProtoHostSpec, MockUpdateSpec](
				hostService, tc.planHosts, tc.stateHosts,
			)

			assert.Equal(t, tc.expectedToCreate, toCreate, "Unexpected hosts to create.")
			assert.Equal(t, tc.expectedToUpdate, toUpdate, "Unexpected hosts to update.")
			assert.Equal(t, tc.expectedToDelete, toDelete, "Unexpected hosts to delete.")
			assert.False(t, diags.HasError(), "Unexpected diagnostics with errors.")
		})
	}
}

func TestShardsDiff(t *testing.T) {
	tests := []struct {
		name             string
		planHosts        map[string]MockHost
		stateHosts       map[string]MockHost
		expectedToCreate map[string][]MockHost
		expectedToDelete map[string][]MockHost
		expectDiags      bool
	}{
		{
			name: "Basic Shard Comparison",
			planHosts: map[string]MockHost{
				"host1": {Shard: "shard1"},
				"host2": {Shard: "shard2"},
			},
			stateHosts: map[string]MockHost{
				"host2": {Shard: "shard2"},
				"host3": {Shard: "shard3"},
			},
			expectedToCreate: map[string][]MockHost{
				"shard1": {{Shard: "shard1"}},
			},
			expectedToDelete: map[string][]MockHost{
				"shard3": {{Shard: "shard3"}},
			},
			expectDiags: false,
		},
		{
			name: "Empty Shard Name Single",
			planHosts: map[string]MockHost{
				"host1": {Shard: ""},
			},
			stateHosts: map[string]MockHost{
				"host2": {Shard: ""},
			},
			expectedToCreate: nil,
			expectedToDelete: nil,
			expectDiags:      false,
		},
		{
			name: "Empty Shard Name Single2",
			planHosts: map[string]MockHost{
				"host1": {Shard: ""},
			},
			stateHosts: map[string]MockHost{
				"host2": {Shard: "shard1"},
			},
			expectedToCreate: nil,
			expectedToDelete: nil,
			expectDiags:      false,
		},
		{
			name: "Empty Shard Name Error",
			planHosts: map[string]MockHost{
				"host1": {Shard: ""},
				"host2": {Shard: "shard1"},
			},
			stateHosts: map[string]MockHost{
				"host3": {Shard: "shard2"},
			},
			expectedToCreate: nil,
			expectedToDelete: nil,
			expectDiags:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			toCreateShards, toDeleteShards, diags := shardsDiff(tc.planHosts, tc.stateHosts)

			assert.Equal(t, tc.expectedToCreate, toCreateShards, "Mismatched shards to create.")
			assert.Equal(t, tc.expectedToDelete, toDeleteShards, "Mismatched shards to delete.")
			assert.Equal(t, tc.expectDiags, diags.HasError(), "Unexpected diagnostics state.")
		})
	}
}

func TestDeleteHostsDependsOnShards(t *testing.T) {
	tests := []struct {
		name           string
		toCreateHosts  []MockHost
		toDeleteHosts  []string
		toCreateShards map[string][]MockHost
		toDeleteShards map[string][]MockHost
		expectedResC   []MockHost
		expectedResD   []string
	}{
		{
			name: "Basic Deletion",
			toCreateHosts: []MockHost{
				{Shard: "shard1"},
				{Shard: "shard2"},
			},
			toDeleteHosts: []string{"host1.example.com", "host2.example.com"},
			toCreateShards: map[string][]MockHost{
				"shard1": {{FQDN: "host3.example.com"}},
			},
			toDeleteShards: map[string][]MockHost{
				"shard3": {{FQDN: "host2.example.com"}},
			},
			expectedResC: []MockHost{{Shard: "shard2"}},
			expectedResD: []string{"host1.example.com"},
		},
		{
			name: "No Changes",
			toCreateHosts: []MockHost{
				{Shard: "shard3"},
			},
			toDeleteHosts: []string{"host4.example.com"},
			toCreateShards: map[string][]MockHost{
				"shard4": {{FQDN: "host5.example.com"}},
			},
			toDeleteShards: map[string][]MockHost{
				"shard5": {{FQDN: "host6.example.com"}},
			},
			expectedResC: []MockHost{{Shard: "shard3"}},
			expectedResD: []string{"host4.example.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resC, resD := deleteHostsDependsOnShards(tc.toCreateHosts, tc.toDeleteHosts, tc.toCreateShards, tc.toDeleteShards)
			assert.ElementsMatch(t, tc.expectedResC, resC, "Unexpected result for hosts to create.")
			assert.ElementsMatch(t, tc.expectedResD, resD, "Unexpected result for hosts to delete.")
		})
	}
}
