package mdb_sharded_postgresql_shard

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
)

func TestYandexProvider_MDBSPQRShardSpecExpand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname string
		spec     types.Object
		expected spqr.ShardSpec_Spec
	}{
		{
			"MDBPostgreSQLSpec",
			types.ObjectValueMust(shardSpecType.AttrTypes, map[string]attr.Value{
				"mdb_postgresql": types.StringValue("cid1"),
			}),
			&spqr.ShardSpec_MdbPostgresql{
				MdbPostgresql: &spqr.MDBPostgreSQL{
					ClusterId: "cid1",
				},
			},
		},
	}

	for _, tc := range cases {
		output, diags := expandShardSpec(ctx, tc.spec)
		if diags.HasError() {
			t.Errorf("Unexpected expand diagnostics status %s test: errors: %v", tc.testname, diags.Errors())
			continue
		}

		if !reflect.DeepEqual(output, tc.expected) {
			t.Errorf(
				"Unexpected expand result value %s test: expected %s, actual %s",
				tc.testname,
				tc.expected,
				output,
			)
		}
	}
}
