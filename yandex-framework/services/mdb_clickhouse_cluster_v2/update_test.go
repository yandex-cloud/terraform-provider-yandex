package mdb_clickhouse_cluster_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
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

			state := &models.Cluster{}
			diags := c.stateVal.As(ctx, state, datasize.DefaultOpts)
			if diags.HasError() {
				t.Fatalf(
					"Unexpected diagnostics in As() for %s: %v",
					c.testname,
					diags.Errors(),
				)
			}

			plan := &models.Cluster{}
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
