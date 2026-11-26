package mdb_clickhouse_cluster_v2

import (
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouse "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_cluster_v2/models"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_cluster_v2/utils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestYandexProvider_MDBClickHouseClusterPrepareUpdateRequests(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname string
		stateVal types.Object
		planVal  types.Object

		expectMoveReq          bool
		expectVersionReq       bool
		expectClusterReq       bool
		expectFormatSchemasReq bool
		expectMlModelReq       bool
		expectShardGroupReq    bool
	}{
		{
			testname: "MaximalChangesCheck",
			stateVal: minimalConfig,
			planVal:  maximalConfig,

			expectMoveReq:          true,
			expectVersionReq:       true,
			expectClusterReq:       true,
			expectFormatSchemasReq: true,
			expectMlModelReq:       true,
			expectShardGroupReq:    true,
		},
		{
			testname: "NoChangesCheck",
			stateVal: minimalConfig,
			planVal:  minimalConfig,

			expectMoveReq:          false,
			expectVersionReq:       false,
			expectClusterReq:       false,
			expectFormatSchemasReq: false,
			expectMlModelReq:       false,
			expectShardGroupReq:    false,
		},
	}

	for _, c := range cases {
		t.Run(c.testname, func(t *testing.T) {
			t.Parallel()

			state := &models.ClusterResource{}
			diags := c.stateVal.As(ctx, state, datasize.DefaultOpts)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in As() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			plan := &models.ClusterResource{}
			diags = c.planVal.As(ctx, plan, datasize.DefaultOpts)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in As() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			// Check folder update
			moveReq := prepareFolderIdUpdateRequest(state, plan)
			if c.expectMoveReq && moveReq == nil {
				t.Errorf("Expected MoveClusterRequest, got nil in %s", c.testname)
			}
			if !c.expectMoveReq && moveReq != nil {
				t.Errorf("Did not expect MoveClusterRequest, got non-nil in %s", c.testname)
			}

			// Check version update
			versionReq := prepareVersionUpdateRequest(state, plan)
			if c.expectVersionReq && versionReq == nil {
				t.Errorf("Expected Version UpdateClusterRequest, got nil in %s", c.testname)
			}
			if !c.expectVersionReq && versionReq != nil {
				t.Errorf("Did not expect Version UpdateClusterRequest, got non-nil in %s", c.testname)
			}

			// Check cluster update
			clusterReq := prepareClusterUpdateRequest(ctx, state, plan, "", false, &diags)
			if diags.HasError() {
				t.Errorf("Unexpected diagnostics in prepareClusterUpdateRequest for %s: %v", c.testname, diags.Errors())
				return
			}

			if c.expectClusterReq {
				if clusterReq == nil {
					t.Errorf("Expected UpdateClusterRequest, got nil in %s", c.testname)
				}

				if len(clusterReq.GetUpdateMask().GetPaths()) == 0 {
					t.Errorf("Expected non-empty UpdateMask.Paths in UpdateClusterRequest for %s", c.testname)
				}
			} else {
				if clusterReq != nil {
					t.Errorf("Did not expect UpdateClusterRequest, got non-nil in %s", c.testname)
				}
			}

			// Check format schema update
			stateFormatSchemas := models.ExpandListFormatSchema(ctx, state.FormatSchema, clusterId, &diags)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in ExpandListFormatSchema() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			fsDelete, fsUpdate, fsCreate := prepareFormatSchemaUpdateRequests(ctx, stateFormatSchemas, plan, &diags)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in prepareFormatSchemaUpdateRequests() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			if c.expectFormatSchemasReq && len(fsDelete) == 0 && len(fsUpdate) == 0 && len(fsCreate) == 0 {
				t.Errorf("Expected some format-schema operations in %s, got none", c.testname)
			}
			if !c.expectFormatSchemasReq && (len(fsDelete) != 0 || len(fsUpdate) != 0 || len(fsCreate) != 0) {
				t.Errorf("Did not expect format-schema operations in %s, got del=%v upd=%d add=%d",
					c.testname, len(fsDelete), len(fsUpdate), len(fsCreate))
			}

			// Check ml model update
			stateMlModels := models.ExpandListMLModel(ctx, state.MLModel, clusterId, &diags)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in ExpandListMLModel() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			mlDelete, mlUpdate, mlCreate := prepareMlModelUpdateRequests(ctx, stateMlModels, plan, &diags)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in prepareMlModelUpdateRequests() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			if c.expectFormatSchemasReq && len(mlDelete) == 0 && len(mlUpdate) == 0 && len(mlCreate) == 0 {
				t.Errorf("Expected some ml model operations in %s, got none", c.testname)
			}
			if !c.expectFormatSchemasReq && (len(mlDelete) != 0 || len(mlUpdate) != 0 || len(mlCreate) != 0) {
				t.Errorf("Did not expect ml model operations in %s, got del=%v upd=%d add=%d",
					c.testname, len(mlDelete), len(mlUpdate), len(mlCreate))
			}

			// Check shard group update
			stateShardGroups := models.ExpandListShardGroup(ctx, state.ShardGroup, clusterId, &diags)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in ExpandListShardGroup() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			sgDelete, gsUpdate, sgCreate := prepareShardGroupUpdateRequests(ctx, stateShardGroups, state, plan, nil, &diags)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in prepareShardGroupUpdateRequests() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			if c.expectFormatSchemasReq && len(sgDelete) == 0 && len(gsUpdate) == 0 && len(sgCreate) == 0 {
				t.Errorf("Expected some shard group operations in %s, got none", c.testname)
			}
			if !c.expectFormatSchemasReq && (len(sgDelete) != 0 || len(gsUpdate) != 0 || len(sgCreate) != 0) {
				t.Errorf("Did not expect shard group operations in %s, got del=%v upd=%d add=%d",
					c.testname, len(mlDelete), len(gsUpdate), len(sgCreate))
			}
		})
	}
}

func TestYandexProvider_MDBClickHouseClusterPrepareExternalDictionaryUpdateOps(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	makeCluster := func(dicts types.Map) *models.ClusterResource {
		cfg := minimalConfig.Attributes()
		cfg["external_dictionary"] = dicts
		reqVal := types.ObjectValueMust(models.ClusterResourceAttrTypes, cfg)
		cluster := &models.ClusterResource{}
		diags := reqVal.As(ctx, cluster, datasize.DefaultOpts)
		if diags.HasError() {
			t.Fatalf("Unexpected diagnostics in As(): %v", diags.Errors())
		}
		return cluster
	}

	emptyStructure := &clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure{}
	flatLayout := &clickhouseConfig.ClickhouseConfig_ExternalDictionary_Layout{
		Type: clickhouseConfig.ClickhouseConfig_ExternalDictionary_Layout_FLAT,
	}
	fixed300 := &clickhouseConfig.ClickhouseConfig_ExternalDictionary_FixedLifetime{FixedLifetime: 300}

	httpDict := &clickhouseConfig.ClickhouseConfig_ExternalDictionary{
		Name:      "http_dict",
		Structure: emptyStructure,
		Layout:    flatLayout,
		Lifetime:  fixed300,
		Source: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_{
			HttpSource: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource{
				Url:    "https://example.com/dict",
				Format: "CSV",
			},
		},
	}

	chDict := &clickhouseConfig.ClickhouseConfig_ExternalDictionary{
		Name:      "ch_dict",
		Structure: emptyStructure,
		Layout:    flatLayout,
		Lifetime:  fixed300,
		Source: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_ClickhouseSource_{
			ClickhouseSource: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_ClickhouseSource{
				Db: "default", Table: "cities", Host: "rc1a-ch.mdb.yandexcloud.net",
				Port: 9000, User: "ch_user", Password: "ch_pass",
			},
		},
	}

	mysqlDictFromAPI := &clickhouseConfig.ClickhouseConfig_ExternalDictionary{
		Name:      "mysql_dict",
		Structure: emptyStructure,
		Layout:    flatLayout,
		Lifetime:  fixed300,
		Source: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_{
			MysqlSource: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource{
				Db:    "mydb",
				Table: "cities",
				Port:  3306,
				User:  "mysql_user",
				Replicas: []*clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_Replica{
					{Host: "rc1b-mysql.mdb.yandexcloud.net", Priority: 1, Port: 3306, User: "replica_user"},
					{Host: "rc1d-mysql.mdb.yandexcloud.net", Priority: 2, Port: 3306, User: "replica_user"},
				},
			},
		},
	}

	pgReplicas, pgDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.DictionaryPostgresqlReplicaV2AttrTypes}, []models.DictionaryPostgresqlReplicaV2{
		{Host: types.StringValue("pg"), Priority: types.Int64Value(1), Password: types.StringValue("first")},
		{Host: types.StringValue("pg"), Port: types.Int64Value(5433), Priority: types.Int64Value(2), Password: types.StringValue("second")},
	})
	if pgDiags.HasError() {
		t.Fatalf("Unexpected replica diagnostics: %v", pgDiags)
	}
	pgConfig := dictionaryTestConfig(t, map[string]attr.Value{"replicas": pgReplicas})
	var pgState types.Map
	pgDiags.Append(pgConfig.GetAttribute(ctx, path.Root("external_dictionary"), &pgState)...)
	pgStateAPI := models.ExpandExternalDictionaries(ctx, pgState, &pgDiags)
	if pgDiags.HasError() {
		t.Fatalf("Unexpected PostgreSQL dictionary diagnostics: %v", pgDiags)
	}
	pgCurrent := proto.Clone(pgStateAPI[0]).(*clickhouseConfig.ClickhouseConfig_ExternalDictionary)
	pgCurrent.Layout.AllowReadExpiredKeys = wrapperspb.Bool(false)
	pgCurrent.Layout.AccessToKeyFromAttributes = wrapperspb.Bool(false)
	pgSource := pgCurrent.GetPostgresqlSourceV2()
	pgSource.Replicas[0], pgSource.Replicas[1] = pgSource.Replicas[1], pgSource.Replicas[0]
	for i, replica := range pgSource.Replicas {
		replica.Priority = int64(i + 3)
		replica.Password = ""
	}
	pgPlan := models.FlattenExternalDictionaries(ctx, []*clickhouseConfig.ClickhouseConfig_ExternalDictionary{pgCurrent}, pgState, &pgDiags)
	if pgDiags.HasError() {
		t.Fatalf("Unexpected PostgreSQL flatten diagnostics: %v", pgDiags)
	}

	cases := []struct {
		name            string
		current         []*clickhouseConfig.ClickhouseConfig_ExternalDictionary
		state           types.Map
		plan            types.Map
		expectedDelete  []string
		expectedCreates []*clickhouse.CreateClusterExternalDictionaryRequest
	}{
		{
			name:            "no_change",
			current:         []*clickhouseConfig.ClickhouseConfig_ExternalDictionary{httpDict},
			state:           httpDictTF,
			plan:            httpDictTF,
			expectedDelete:  nil,
			expectedCreates: nil,
		},
		{
			name:            "no_change_with_password",
			current:         []*clickhouseConfig.ClickhouseConfig_ExternalDictionary{mysqlDictFromAPI},
			state:           mysqlDictTF,
			plan:            mysqlDictTF,
			expectedDelete:  nil,
			expectedCreates: nil,
		},
		{
			name:    "postgresql_reordered_replicas_with_changed_priorities",
			current: []*clickhouseConfig.ClickhouseConfig_ExternalDictionary{pgCurrent},
			state:   pgState,
			plan:    pgPlan,
		},
		{
			name:           "add_new_dict",
			current:        nil,
			state:          emptyDictMapTF,
			plan:           httpDictTF,
			expectedDelete: nil,
			expectedCreates: []*clickhouse.CreateClusterExternalDictionaryRequest{
				{ClusterId: clusterId, ExternalDictionary: httpDict},
			},
		},
		{
			name:            "remove_dict",
			current:         []*clickhouseConfig.ClickhouseConfig_ExternalDictionary{httpDict},
			state:           httpDictTF,
			plan:            emptyDictMapTF,
			expectedDelete:  []string{"http_dict"},
			expectedCreates: nil,
		},
		{
			name:           "modify_dict",
			state:          httpDictTF,
			current:        []*clickhouseConfig.ClickhouseConfig_ExternalDictionary{httpDict},
			plan:           httpDictModifiedTF,
			expectedDelete: []string{"http_dict"},
			expectedCreates: []*clickhouse.CreateClusterExternalDictionaryRequest{
				{
					ClusterId: clusterId,
					ExternalDictionary: &clickhouseConfig.ClickhouseConfig_ExternalDictionary{
						Name:      "http_dict",
						Structure: emptyStructure,
						Layout:    flatLayout,
						Lifetime:  fixed300,
						Source: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_{
							HttpSource: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource{
								Url: "https://example.com/dict", Format: "TSV",
							},
						},
					},
				},
			},
		},
		{
			name:           "mixed_add_remove",
			current:        []*clickhouseConfig.ClickhouseConfig_ExternalDictionary{httpDict},
			state:          httpDictTF,
			plan:           makeDictMapTF(map[string]types.Object{"ch_dict": makeDictTF(chDictSourceTF)}),
			expectedDelete: []string{"http_dict"},
			expectedCreates: []*clickhouse.CreateClusterExternalDictionaryRequest{
				{ClusterId: clusterId, ExternalDictionary: chDict},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			stateCluster := makeCluster(c.state)
			stateDicts := models.ExpandExternalDictionaries(ctx, stateCluster.ExternalDictionary, &diags)
			if diags.HasError() {
				t.Fatalf("Unexpected diagnostics expanding state: %v", diags.Errors())
			}
			planCluster := makeCluster(c.plan)
			toDelete, toCreate := prepareExternalDictionaryUpdateOps(ctx, clusterId, c.current, stateDicts, planCluster, &diags)
			if diags.HasError() {
				t.Fatalf("Unexpected diagnostics: %v", diags.Errors())
			}

			sort.Strings(toDelete)
			sort.Strings(c.expectedDelete)
			if len(toDelete) != len(c.expectedDelete) {
				t.Fatalf("Expected deletes %v, got %v", c.expectedDelete, toDelete)
			}
			for i, name := range c.expectedDelete {
				if toDelete[i] != name {
					t.Errorf("Expected delete[%d]=%q, got %q", i, name, toDelete[i])
				}
			}

			if len(toCreate) != len(c.expectedCreates) {
				t.Fatalf("Expected %d creates, got %d", len(c.expectedCreates), len(toCreate))
			}
			for i, expected := range c.expectedCreates {
				utils.AssertProtoEqual(t, c.name, expected, toCreate[i])
			}
		})
	}
}

func TestYandexProvider_MDBClickHouseClusterExternalReplicaUpdates(t *testing.T) {
	t.Parallel()
	key := externalReplicaKey{"analytics", "external", "host"}
	cases := []struct {
		name             string
		state            models.ExternalShardReplica
		plan             models.ExternalShardReplica
		priority         *wrapperspb.Int64Value
		isNew            bool
		expectedPaths    []string
		expectedPassword string
		expectedError    bool
	}{
		{
			name:  "unchanged_version",
			state: models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(1)},
			plan:  models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(1), PasswordWo: types.StringValue("ignored")},
		},
		{
			name:     "priority_preserves_password",
			state:    models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(1)},
			plan:     models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(1), PasswordWo: types.StringValue("ignored")},
			priority: wrapperspb.Int64(10), expectedPaths: []string{"priority"},
		},
		{
			name:          "rotate_write_only",
			state:         models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(1)},
			plan:          models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(2), PasswordWo: types.StringValue("new-secret")},
			expectedPaths: []string{"password"}, expectedPassword: "new-secret",
		},
		{
			name:          "missing_write_only_password",
			state:         models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(1)},
			plan:          models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(2)},
			expectedError: true,
		},
		{
			name:  "new_replica",
			plan:  models.ExternalShardReplica{PasswordWoVersion: types.Int64Value(1), PasswordWo: types.StringValue("new-secret")},
			isNew: true, expectedPaths: []string{"user", "password"}, expectedPassword: "new-secret",
		},
		{
			name:          "rotate_legacy",
			state:         models.ExternalShardReplica{Password: types.StringValue("old-secret")},
			plan:          models.ExternalShardReplica{Password: types.StringValue("new-secret")},
			expectedPaths: []string{"password"}, expectedPassword: "new-secret",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			current := []*clickhouse.ExternalShard_Replica{{Host: "other"}}
			if !c.isNew {
				current = append([]*clickhouse.ExternalShard_Replica{{Host: "host"}}, current...)
			}
			target := []*clickhouse.ExternalShard_Replica{{Host: "host", Priority: c.priority}, {Host: "other"}}
			var diags diag.Diagnostics
			replicas, changed := externalReplicaUpdateSpecs("analytics", "external", current, target,
				externalReplicaMap{key: c.state}, externalReplicaMap{key: c.plan}, &diags)
			if diags.HasError() != c.expectedError {
				t.Fatalf("Expected error %t, got diagnostics: %v", c.expectedError, diags.Errors())
			}
			if c.expectedError {
				return
			}
			if changed != (len(c.expectedPaths) > 0) {
				t.Errorf("Unexpected changed flag: %t", changed)
			}
			expected := &clickhouse.UpdateShardGroupSpec_ExternalShard{
				ReplicaSpecs: []*clickhouse.UpdateShardGroupSpec_ExternalShard_Replica{
					{Host: "host", Priority: c.priority, Password: c.expectedPassword, UpdateMask: &fieldmaskpb.FieldMask{Paths: c.expectedPaths}},
					{Host: "other", UpdateMask: &fieldmaskpb.FieldMask{}},
				},
			}
			utils.AssertProtoEqual(t, c.name, expected, &clickhouse.UpdateShardGroupSpec_ExternalShard{ReplicaSpecs: replicas})
		})
	}
}

func TestYandexProvider_MDBClickHouseClusterShardGroupRefresh(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cases := []struct {
		name    string
		replica models.ExternalShardReplica
	}{
		{
			name:    "write_only",
			replica: models.ExternalShardReplica{Host: types.StringValue("host"), PasswordWoVersion: types.Int64Value(2)},
		},
		{
			name:    "legacy",
			replica: models.ExternalShardReplica{Host: types.StringValue("host"), Password: types.StringValue("legacy-secret")},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			state := testShardGroupCluster(t, testExternalShard(t, "external", c.replica))
			current := []*clickhouse.ShardGroup{{Name: "analytics", ExternalShards: []*clickhouse.ExternalShard{{
				Name: "external", Replicas: []*clickhouse.ExternalShard_Replica{{Host: "host", Password: "must-not-leak"}},
			}}}}
			refreshed := models.FlattenListShardGroup(ctx, current, state.ShardGroup, &diags)
			key := externalReplicaKey{"analytics", "external", "host"}
			replica := externalReplicasByIdentity(ctx, refreshed, nil, &diags)[key]
			if diags.HasError() {
				t.Fatalf("Unexpected diagnostics: %v", diags.Errors())
			}
			if !replica.Password.Equal(c.replica.Password) || !replica.PasswordWo.IsNull() || !replica.PasswordWoVersion.Equal(c.replica.PasswordWoVersion) {
				t.Errorf("Refresh must preserve the state password/version and discard the API password")
			}
			shardNames := refreshed.Elements()[0].(types.Object).Attributes()["shard_names"]
			if !shardNames.Equal(types.ListValueMust(types.StringType, []attr.Value{})) {
				t.Errorf("Expected empty shard_names, got %v", shardNames)
			}
		})
	}
}

func testExternalReplicaPlan(t *testing.T, plan *models.ClusterResource, passwords externalReplicaMap) externalReplicaMap {
	t.Helper()
	var diags diag.Diagnostics
	replicas := externalReplicasByIdentity(context.Background(), plan.ShardGroup, passwords, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	return replicas
}

func testExternalReplica(host string) models.ExternalShardReplica {
	return models.ExternalShardReplica{
		Host: types.StringValue(host), Port: types.Int64Value(9440), Secure: types.BoolValue(true),
		User: types.StringValue("reader"), Priority: types.Int64Value(0), PasswordWoVersion: types.Int64Value(1),
	}
}

func testExternalShard(t *testing.T, name string, replicas ...models.ExternalShardReplica) models.ExternalShardModel {
	t.Helper()
	value, diags := types.ListValueFrom(context.Background(), types.ObjectType{AttrTypes: models.ExternalShardReplicaAttrTypes}, replicas)
	if diags.HasError() {
		t.Fatal(diags)
	}
	return models.ExternalShardModel{Name: types.StringValue(name), Weight: types.Int64Value(100), Replica: value}
}

func testShardGroupCluster(t *testing.T, shards ...models.ExternalShardModel) *models.ClusterResource {
	t.Helper()
	externalShards, diags := types.ListValueFrom(context.Background(), types.ObjectType{AttrTypes: models.ExternalShardAttrTypes}, shards)
	if diags.HasError() {
		t.Fatal(diags)
	}
	groups, diags := types.ListValueFrom(context.Background(), types.ObjectType{AttrTypes: models.ShardGroupAttrTypes}, []models.ShardGroup{{
		Name: types.StringValue("analytics"), Description: types.StringValue("test"),
		ShardNames: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("shard1")}), ExternalShards: externalShards,
	}})
	if diags.HasError() {
		t.Fatal(diags)
	}
	return &models.ClusterResource{Id: types.StringValue("cluster-id"), ShardGroup: groups}
}
