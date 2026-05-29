package mdb_clickhouse_cluster_v2

import (
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouse "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
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
			clusterReq := prepareClusterUpdateRequest(ctx, state, plan, &diags)
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

			sgDelete, gsUpdate, sgCreate := prepareShardGroupUpdateRequests(ctx, stateShardGroups, plan, &diags)
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
